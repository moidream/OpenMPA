import { expect, it, vi } from 'vitest';
import type { UsageRequestVM } from '../src/client/desktop/models';

const api = vi.hoisted(() => ({ onUsage: vi.fn() }));
vi.mock('../src/client/desktop/generated/mygo', () => ({
  Artifact: {},
  Host: {},
  Platform: {},
  Connection: {},
  Workspace: {},
  Deployment: {},
  Schedule: {},
  Channels: {},
  SessionBot: {},
  Session: {},
  events: { usageRequested: { on: api.onUsage } },
}));
import { createDesktopServices } from '../src/client/desktop/services';

it('forwards the typed native usage payload and subscription cleanup unchanged', () => {
  let emit!: (request: UsageRequestVM) => void;
  const remove = vi.fn();
  api.onUsage.mockImplementation((listener: typeof emit) => {
    emit = listener;
    return remove;
  });
  const listener = vi.fn();
  const unsubscribe = createDesktopServices().host!.onUsage!(listener);
  const request = {
    connectionEpoch: 2,
    sessionId: 'selected',
    text: '@member',
  };
  emit(request);
  expect(listener).toHaveBeenCalledExactlyOnceWith(request);
  expect(unsubscribe).toBe(remove);
  unsubscribe();
  expect(remove).toHaveBeenCalledOnce();
});
