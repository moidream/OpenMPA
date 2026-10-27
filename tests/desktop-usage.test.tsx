import { StrictMode } from 'react';
import { act, create, type ReactTestRenderer } from 'react-test-renderer';
import { afterEach, expect, it, vi } from 'vitest';
import { DesktopApp } from '../src/client/desktop/DesktopApp';
import type {
  ConnectionVM,
  DesktopServices,
  NativeShellState,
  UsageRequestVM,
} from '../src/client/desktop/models';

const connection: ConnectionVM = {
  epoch: 2,
  scope: 'scope',
  runtimeId: 'runtime',
  status: 'connected',
};
const sessions = [
  { id: 'first', title: 'First conversation' },
  { id: 'second', title: 'Second conversation' },
];
const renderers: ReactTestRenderer[] = [];

afterEach(async () => {
  await act(async () => {
    for (const renderer of renderers.splice(0)) renderer.unmount();
  });
  vi.unstubAllGlobals();
});

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason: Error) => void;
  const promise = new Promise<T>((done, fail) => {
    resolve = done;
    reject = fail;
  });
  return { promise, resolve, reject };
}

async function mount(
  overrides: Partial<DesktopServices['sessions']> = {},
  strict = false,
  initialStatus: DesktopServices['connection']['status'] = async () =>
    connection,
) {
  vi.stubGlobal('IS_REACT_ACT_ENVIRONMENT', true);
  let navigate!: (action: string, id?: string) => void;
  const usageListeners = new Set<(request: UsageRequestVM) => void>();
  const removeUsage = vi.fn();
  const status = vi.fn(initialStatus);
  const submit = vi.fn(async () => ({
    invocationId: 'invocation',
    status: 'accepted',
  }));
  const read = vi.fn(async (_epoch: number, id: string) => ({
    session: sessions.find((session) => session.id === id)!,
    messages: [],
  }));
  const setShellState = vi.fn(async (_state: NativeShellState) => {});
  const services = {
    host: {
      onNavigation: (listener: typeof navigate) => {
        navigate = listener;
        return () => {};
      },
      onUsage: (listener: (request: UsageRequestVM) => void) => {
        usageListeners.add(listener);
        return () => {
          usageListeners.delete(listener);
          removeUsage();
        };
      },
      setShellState,
      onClose: () => () => {},
      closeState: async () => ({ pending: false, requestId: 0 }),
    },
    workspace: { spaces: async () => [] },
    connection: { status },
    sessions: {
      list: async () => sessions,
      pending: async () => [],
      read,
      submit,
      ...overrides,
    },
  } as unknown as DesktopServices;
  let renderer!: ReactTestRenderer;
  await act(async () => {
    const app = <DesktopApp services={services} />;
    renderer = create(strict ? <StrictMode>{app}</StrictMode> : app);
  });
  renderers.push(renderer);
  const emit = (text: string, sessionId = 'first', connectionEpoch = 2) => {
    for (const listener of usageListeners)
      listener({ text, sessionId, connectionEpoch });
  };
  const select = async (id = 'first') => {
    await act(async () => navigate('selectSession', id));
  };
  const type = async (value: string) => {
    await act(async () =>
      renderer.root
        .findByType('textarea')
        .props.onChange({ target: { value } }),
    );
  };
  const draft = () => renderer.root.findByType('textarea').props.value;
  return {
    renderer,
    navigate,
    emit,
    select,
    type,
    draft,
    read,
    submit,
    status,
    setShellState,
    usageListeners,
    removeUsage,
  };
}

it.each([
  ['resources', 'Resources'],
  ['collaboration', 'Collaboration'],
])(
  'renders native %s and prepends A2A usage once without sending',
  async (view, title) => {
    const app = await mount({}, true);
    await app.select();
    await app.type('Keep my unsent draft  ');
    await act(async () => app.navigate(view));
    expect(app.renderer.root.findAllByType('textarea')).toHaveLength(0);
    expect(
      app.renderer.root.findByProps({ 'aria-label': `Native ${view} content` }),
    ).toBeDefined();
    expect(app.setShellState).toHaveBeenLastCalledWith(
      expect.objectContaining({
        view,
        pageTitle: title,
        selectedSession: 'first',
      }),
    );
    expect(app.usageListeners.size).toBe(1);
    await act(async () => app.emit('@selected task'));
    expect(app.draft()).toBe('@selected task\nKeep my unsent draft  ');
    expect(app.setShellState).toHaveBeenLastCalledWith(
      expect.objectContaining({ view: 'chat', selectedSession: 'first' }),
    );
    await act(async () => app.navigate(view));
    await act(async () => app.navigate('home'));
    expect(app.draft()).toBe('@selected task\nKeep my unsent draft  ');
    expect(app.submit).not.toHaveBeenCalled();
    await act(async () => app.renderer.unmount());
    expect(app.usageListeners.size).toBe(0);
    expect(app.removeUsage).toHaveBeenCalled();
  },
);

it('preserves separate drafts across session switches and rejects stale usage in the same event batch', async () => {
  const app = await mount();
  await app.select();
  await app.type('First draft');
  await act(async () => {
    app.navigate('selectSession', 'second');
    app.emit('wrong session');
    app.emit('@second task', 'second');
  });
  expect(app.draft()).toBe('@second task');
  await app.type('Second draft');
  await app.select();
  expect(app.draft()).toBe('First draft');
  await act(async () => {
    app.emit('/skill');
    app.emit('@colleague task');
  });
  expect(app.draft()).toBe('@colleague task\nFirst draft\n/skill');
  await app.select('second');
  expect(app.draft()).toBe('Second draft');
  expect(app.submit).not.toHaveBeenCalled();
});

it('ignores wrong epochs, missing sessions and blank text without navigating', async () => {
  const app = await mount();
  await act(async () => {
    app.navigate('resources');
    app.emit('no selected session');
  });
  expect(app.renderer.root.findAllByType('textarea')).toHaveLength(0);
  await app.select();
  await app.type('Unsent');
  await act(async () => app.navigate('collaboration'));
  await act(async () => {
    app.emit('old connection', 'first', 1);
    app.emit('future connection', 'first', 3);
    app.emit('unknown session', 'missing');
    app.emit('other session', 'second');
    app.emit(' \n\t ');
  });
  expect(app.setShellState).toHaveBeenLastCalledWith(
    expect.objectContaining({ view: 'collaboration' }),
  );
  await act(async () => app.navigate('home'));
  expect(app.draft()).toBe('Unsent');
  expect(app.submit).not.toHaveBeenCalled();
});

it('discards drafts immediately on reset and never revives them from a reordered status response', async () => {
  const app = await mount();
  await app.select();
  await app.type('Old account draft');
  await act(async () => app.emit('/old skill'));
  expect(app.draft()).toBe('Old account draft\n/old skill');
  await act(async () => app.navigate('resources'));
  const older = deferred<ConnectionVM>();
  const latest = deferred<ConnectionVM>();
  app.status.mockImplementationOnce(() => older.promise);
  app.status.mockImplementationOnce(() => latest.promise);
  await act(async () => {
    app.navigate('connectionChanged');
    app.emit('late old usage');
    app.navigate('connectionChanged');
  });
  await act(async () => latest.resolve({ ...connection, epoch: 4 }));
  await app.select();
  expect(app.draft()).toBe('');
  await act(async () => app.emit('/new skill', 'first', 4));
  expect(app.draft()).toBe('/new skill');
  await act(async () => older.resolve({ ...connection, epoch: 3 }));
  expect(app.draft()).toBe('/new skill');
  expect(app.setShellState).toHaveBeenLastCalledWith(
    expect.objectContaining({ connectionEpoch: 4, selectedSession: 'first' }),
  );
});

it('keeps a fresh usage draft when delayed pending-run recovery resolves', async () => {
  const history =
    deferred<Awaited<ReturnType<DesktopServices['sessions']['read']>>>();
  const app = await mount({
    read: () => history.promise,
    pending: async () => [
      {
        sessionId: 'first',
        operationId: 'operation',
        idempotencyKey: 'key',
        content: 'Older submitted prompt',
        state: 'unknown',
      },
    ],
  });
  await app.select();
  await app.type('New unsent draft');
  await act(async () => app.navigate('resources'));
  await act(async () => app.emit('/skill'));
  await act(async () =>
    history.resolve({ session: sessions[0], messages: [] }),
  );
  expect(app.draft()).toBe('New unsent draft\n/skill');
  expect(app.renderer.root.findByProps({ type: 'submit' }).props.disabled).toBe(
    true,
  );
  expect(app.submit).not.toHaveBeenCalled();
});

it('does not resurrect a deliberately cleared pending draft after visiting native pages', async () => {
  const app = await mount({
    pending: async () => [
      {
        sessionId: 'first',
        operationId: 'operation',
        idempotencyKey: 'key',
        content: 'Original prompt',
        state: 'unknown',
      },
    ],
  });
  await app.select();
  expect(app.draft()).toBe('Original prompt');
  await app.type('');
  await act(async () => app.navigate('collaboration'));
  await act(async () => app.navigate('home'));
  expect(app.draft()).toBe('');
});

it.each([
  { usage: ['@next task'], expected: '@next task' },
  { usage: ['/skill'], expected: '/skill' },
  {
    usage: ['@next task', '/skill', '@last task'],
    expected: '@last task\n@next task\n/skill',
  },
])(
  'clears only sent text after a receipt with new usage $usage',
  async ({ usage, expected }) => {
    const receipt = deferred<{ invocationId: string; status: string }>();
    const submit = vi.fn(() => receipt.promise);
    const app = await mount({
      submit,
      subscribe: () => ({
        events: (async function* () {})(),
        finished: Promise.resolve(),
        close() {},
      }),
    });
    await app.select();
    await app.type('Submitted draft');
    await act(async () => {
      app.renderer.root
        .findByType('form')
        .props.onSubmit({ preventDefault() {} });
    });
    await act(async () => {
      for (const text of usage) app.emit(text);
    });
    await act(async () =>
      receipt.resolve({ invocationId: 'invocation', status: 'accepted' }),
    );
    expect(app.draft()).toBe(expected);
    expect(submit).toHaveBeenCalledTimes(1);
    expect(submit).toHaveBeenCalledWith(
      2,
      'first',
      'Submitted draft',
      undefined,
    );
  },
);

it.each([
  ['NUL', '@selected task\0hidden'],
  ['ASCII overflow by one byte', '@selected ' + 'a'.repeat(64 * 1024 - 9)],
  ['UTF-8 overflow', '@selected ' + '\u00e9'.repeat(32 * 1024)],
  ['four-byte UTF-8 overflow', '@selected ' + '\ud83d\ude00'.repeat(16 * 1024)],
])(
  'rejects %s usage without navigating or changing the draft',
  async (_name, text) => {
    const app = await mount();
    await app.select();
    await app.type('Unsent');
    await act(async () => app.navigate('collaboration'));
    await act(async () => app.emit(text));
    expect(app.renderer.root.findAllByType('textarea')).toHaveLength(0);
    await act(async () => app.navigate('home'));
    expect(app.draft()).toBe('Unsent');
    expect(app.submit).not.toHaveBeenCalled();
  },
);

it.each([
  ['ASCII limit', '@selected ' + 'a'.repeat(64 * 1024 - 10)],
  ['UTF-8 limit', '@selected ' + '\u00e9'.repeat((64 * 1024 - 10) / 2)],
  [
    'four-byte UTF-8 limit',
    '@selected ' + '\ud83d\ude00'.repeat((64 * 1024 - 12) / 4) + 'ab',
  ],
  ['multiline task', '@selected first line\nsecond line\r\nthird line'],
  ['quoted member', '@"Selected member" first line\nsecond line'],
])('preserves a valid %s before the existing draft', async (_name, text) => {
  const app = await mount();
  await app.select();
  await app.type('Existing draft\n');
  await act(async () => app.emit(text));
  expect(app.draft()).toBe(`${text}\nExisting draft\n`);
  expect(app.submit).not.toHaveBeenCalled();
});

it.each(['Draft', 'Draft ', 'Draft\n', ''])(
  'keeps Skill append behavior for draft %j',
  async (draft) => {
    const app = await mount();
    await app.select();
    await app.type(draft);
    await act(async () => app.emit('/skill'));
    expect(app.draft()).toBe(
      `${draft}${draft && !/\s$/.test(draft) ? '\n' : ''}/skill`,
    );
  },
);

it.each(['@next task\nSubmitted draft, with manual edits', ''])(
  'keeps manual edits %j when a receipt arrives after native usage',
  async (edited) => {
    const receipt = deferred<{ invocationId: string; status: string }>();
    const app = await mount({
      submit: () => receipt.promise,
      subscribe: () => ({
        events: (async function* () {})(),
        finished: Promise.resolve(),
        close() {},
      }),
    });
    await app.select();
    await app.type('Submitted draft');
    await act(async () => {
      app.renderer.root
        .findByType('form')
        .props.onSubmit({ preventDefault() {} });
    });
    await act(async () => app.emit('@next task'));
    await app.type(edited);
    await act(async () => app.emit('/skill'));
    await act(async () =>
      receipt.resolve({ invocationId: 'invocation', status: 'accepted' }),
    );
    expect(app.draft()).toBe(edited ? `${edited}\n/skill` : '/skill');
  },
);

it('clears pre-submit usage along with the sent draft, retaining only later choices', async () => {
  const receipt = deferred<{ invocationId: string; status: string }>();
  const submit = vi.fn(() => receipt.promise);
  const app = await mount({
    submit,
    subscribe: () => ({
      events: (async function* () {})(),
      finished: Promise.resolve(),
      close() {},
    }),
  });
  await app.select();
  await app.type('Submitted draft');
  await act(async () => app.emit('@selected task'));
  await act(async () => {
    app.renderer.root
      .findByType('form')
      .props.onSubmit({ preventDefault() {} });
  });
  await act(async () => app.emit('@next task'));
  await act(async () =>
    receipt.resolve({ invocationId: 'invocation', status: 'accepted' }),
  );
  expect(app.draft()).toBe('@next task');
  expect(submit).toHaveBeenCalledWith(
    2,
    'first',
    '@selected task\nSubmitted draft',
    undefined,
  );
});

it('keeps accepted usage in its original session when selection changes later in the same batch', async () => {
  const app = await mount();
  await app.select();
  await app.type('First draft');
  await act(async () => {
    app.emit('@selected task');
    app.navigate('selectSession', 'second');
    app.emit('@stale task');
    app.emit('/skill', 'second');
  });
  expect(app.draft()).toBe('/skill');
  await app.select();
  expect(app.draft()).toBe('@selected task\nFirst draft');
});

it('discards usage queued before a connection reset in the same batch', async () => {
  const app = await mount();
  await app.select();
  await app.type('Old draft');
  app.status.mockResolvedValueOnce({ ...connection, epoch: 3 });
  await act(async () => {
    app.emit('@selected task');
    app.navigate('connectionChanged');
    app.emit('@stale task');
  });
  await app.select();
  expect(app.draft()).toBe('');
  await act(async () => app.emit('@next task', 'first', 3));
  expect(app.draft()).toBe('@next task');
});

it('ignores usage until the new connection session list has loaded', async () => {
  const nextList = deferred<typeof sessions>();
  const list = vi.fn(async () => sessions);
  const app = await mount({ list });
  await app.select();
  await app.type('Old draft');
  await act(async () => app.navigate('collaboration'));
  list.mockImplementationOnce(() => nextList.promise);
  app.status.mockResolvedValueOnce({ ...connection, epoch: 3 });
  await act(async () => app.navigate('connectionChanged'));
  await act(async () => {
    app.navigate('selectSession', 'first');
    app.emit('@too-early task', 'first', 3);
  });
  expect(app.renderer.root.findAllByType('textarea')).toHaveLength(0);
  await act(async () => nextList.resolve(sessions));
  await app.select();
  expect(app.draft()).toBe('');
  await act(async () => {
    app.emit('@stale task');
    app.emit('@next task', 'first', 3);
  });
  expect(app.draft()).toBe('@next task');
});

it('does not overwrite the current connection with a late initial status response', async () => {
  const initial = deferred<ConnectionVM>();
  const app = await mount({}, true, () => initial.promise);
  app.status.mockResolvedValueOnce({ ...connection, epoch: 3 });
  await act(async () => app.navigate('connectionChanged'));
  await app.select();
  await act(async () => app.emit('@selected task', 'first', 3));
  await act(async () => initial.resolve(connection));
  expect(app.draft()).toBe('@selected task');
  expect(app.setShellState).toHaveBeenLastCalledWith(
    expect.objectContaining({ connectionEpoch: 3, selectedSession: 'first' }),
  );
});

it('keeps the composed draft when submission fails instead of returning a receipt', async () => {
  const receipt = deferred<{ invocationId: string; status: string }>();
  const app = await mount({ submit: () => receipt.promise });
  await app.select();
  await app.type('Submitted draft');
  await act(async () => {
    app.renderer.root
      .findByType('form')
      .props.onSubmit({ preventDefault() {} });
  });
  await act(async () => app.emit('@next task'));
  await act(async () => receipt.reject(new Error('Receipt unavailable')));
  expect(app.draft()).toBe('@next task\nSubmitted draft');
});

it('ignores a receipt from the old connection without clearing new usage', async () => {
  const receipt = deferred<{ invocationId: string; status: string }>();
  const app = await mount({ submit: () => receipt.promise });
  await app.select();
  await app.type('Submitted draft');
  await act(async () => {
    app.renderer.root
      .findByType('form')
      .props.onSubmit({ preventDefault() {} });
  });
  app.status.mockResolvedValueOnce({ ...connection, epoch: 3 });
  await act(async () => app.navigate('connectionChanged'));
  await app.select();
  await app.type('New account draft');
  await act(async () => app.emit('@next task', 'first', 3));
  await act(async () =>
    receipt.resolve({ invocationId: 'invocation', status: 'accepted' }),
  );
  expect(app.draft()).toBe('@next task\nNew account draft');
});

it('clears sent text when the receipt arrives after visiting a native page', async () => {
  const receipt = deferred<{ invocationId: string; status: string }>();
  const app = await mount({ submit: () => receipt.promise });
  await app.select();
  await app.type('Submitted draft');
  await act(async () => {
    app.renderer.root
      .findByType('form')
      .props.onSubmit({ preventDefault() {} });
  });
  await act(async () => app.navigate('collaboration'));
  await act(async () => app.emit('@next task'));
  await act(async () =>
    receipt.resolve({ invocationId: 'invocation', status: 'accepted' }),
  );
  expect(app.draft()).toBe('@next task');
});
