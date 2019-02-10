// TODO: Declare stricter types. Consider contributing to machina.js itself [0]
// or DefinitelyTyped [1] (but preferably the former).
//
// [0]: https://github.com/ifandelse/machina.js/issues/158
// [1]: https://github.com/DefinitelyTyped/DefinitelyTyped/issues/25075
declare module 'machina' {
  // export as namespace machina;

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
    on(eventName: string, callback: EventListener): EventOnResult;
    off(eventName?: string, callback?: EventListener): void;
  }

  export interface Client {
    __machina__: ClientInstance;
  }

  export interface ClientInstance {
    targetReplayState: any;
    state: string;
    priorState: string;
    priorAction: string;
    currentAction: string;
    currentActionArgs: any[];
    inputQueue: any[];
    inExitHandler: boolean;
    initialize: () => void;
  }

  export class Fsm implements ClientInstance {
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
    on(eventName: string, callback: EventListener): EventOnResult;
    off(eventName?: string, callback?: EventListener): void;
  }

  export interface FsmOptions {
    initialState?: string;
    eventListeners?: EventListeners;
    states?: States;
    namespace?: string;
    initialize?: () => void;
  }

  export interface States {
    [name: string]: State;
  }

  export interface State {
    _onEnter?: () => void;
    _onExit?: () => void;
    '*'?: () => void;
    [action: string]: string | ((...args: any[]) => void) | undefined;
  }

  export interface EventListeners {
    [eventName: string]: EventListener[] | undefined;
  }

  export type EventListener = (...args: any[]) => void;

  export interface EventOnResult {
    eventName: string;
    callback: EventListener;
    off: () => void;
  }
}
