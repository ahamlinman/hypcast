// TODO: Declare stricter types. Consider contributing to machina.js itself [0]
// or DefinitelyTyped [1] (but preferably the former).
//
// [0]: https://github.com/ifandelse/machina.js/issues/158
// [1]: https://github.com/DefinitelyTyped/DefinitelyTyped/issues/25075
declare module 'machina' {
  // export as namespace machina;

  export const eventListeners: EventListeners;
  function on(eventName: string, callback: EventListener): EventSubscription;
  function off(eventName: string, callback?: EventListener): void;
  function emit(eventName: string, ...args: any[]): void;

  export namespace utils {
    function createUUID(): string;
    function getDefaultClientMeta(): ClientMeta;
    function getDefaultOptions(): DefaultOptions;
    function getLeaklessArgs(): any;
    function listenToChild(): EventSubscription;
    function makeFsmNamespace(): string;
  }

  export class BehavioralFsm {
    new(options: FsmOptions): BehavioralFsm;
    static extend(options: FsmOptions): typeof BehavioralFsm;

    initialState: string;
    eventListeners: EventListeners;
    states: States;
    namespace?: string;
    initialize: () => void;

    emit(eventName: string, ...args: any[]): void;
    handle(client: Client, eventName: string, ...args: any[]): void;
    transition(client: Client, stateName: string): void;
    processQueue(client: Client, type: string): void;
    clearQueue(client: Client, type?: string, stateName?: string): void;
    deferUntilTransition(client: Client, stateName?: string): void;
    deferAndTransition(client: Client, stateName: string): void;
    compositeState(client: Client): string;
    on(eventName: string, callback: EventListener): EventSubscription;
    off(eventName?: string, callback?: EventListener): void;
  }

  export interface Client {
    __machina__?: ClientMeta;
  }

  export interface ClientMeta {
    targetReplayState: any;
    state: string;
    priorState: string;
    priorAction: string;
    currentAction: string;
    currentActionArgs: any[];
    inputQueue: any[];
    inExitHandler: boolean;
    initialize?: () => void;
  }

  export class Fsm implements ClientMeta {
    new(options: FsmOptions): Fsm;
    static extend(options: FsmOptions): typeof Fsm;

    initialState: string;
    eventListeners: EventListeners;
    states: States;
    namespace?: string;
    initialize: () => void;

    targetReplayState: any;
    state: string;
    priorState: string;
    priorAction: string;
    currentAction: string;
    currentActionArgs: any[];
    inputQueue: any[];
    inExitHandler: boolean;

    emit(eventName: string, ...args: any[]): void;
    handle(eventName: string, ...args: any[]): void;
    transition(stateName: string): void;
    processQueue(type: string): void;
    clearQueue(type?: string, stateName?: string): void;
    deferUntilTransition(stateName?: string): void;
    deferAndTransition(stateName: string): void;
    compositeState(): string;
    on(eventName: string, callback: EventListener): EventSubscription;
    off(eventName?: string, callback?: EventListener): void;
  }

  export interface FsmOptions {
    initialState?: string;
    eventListeners?: EventListeners;
    states?: States;
    namespace?: string;
    initialize?: () => void;
  }

  export interface DefaultOptions extends FsmOptions {
    useSafeEmit: boolean;
    hierarchy: any;
    pendingDelegations: any;
  }

  export interface States {
    [name: string]: State;
  }

  export interface State {
    _child?: StateChild;
    _onEnter?: () => void;
    _onExit?: () => void;
    '*'?: () => void;

    [action: string]: StateChild | string | ((...args: any[]) => void) | undefined;
  }

  export type StateChild = Fsm | (() => Fsm) | { factory(): Fsm };

  export interface EventListeners {
    [eventName: string]: EventListener[] | undefined;
  }

  export type EventListener = (...args: any[]) => void;

  export interface EventSubscription {
    eventName: string;
    callback: EventListener;
    off: () => void;
  }
}
