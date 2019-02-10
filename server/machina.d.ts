// TODO: Declare stricter types. Consider contributing to machina.js itself [0]
// or DefinitelyTyped [1] (but preferably the former).
//
// [0]: https://github.com/ifandelse/machina.js/issues/158
// [1]: https://github.com/DefinitelyTyped/DefinitelyTyped/issues/25075
declare module 'machina' {
  // export as namespace machina;

  export interface State {
    _onEnter?: () => void;
    _onExit?: () => void;
    '*'?: () => void;
    [action: string]: string | ((...args: any[]) => void) | undefined;
  }

  export interface States {
    [name: string]: State;
  }

  export type EventListener = (...args: any[]) => void;

  export interface EventListeners {
    [eventName: string]: EventListener[] | undefined;
  }

  export interface EventOnResult {
    eventName: string;
    callback: EventListener;
    off: () => void;
  }

  export interface FsmOptions {
    initialState?: string;
    states?: States;
    eventListeners?: EventListeners;
    namespace?: string;
    initialize?: () => void;
  }

  export class Fsm {
    new(options: FsmOptions): Fsm;
    static extend(options: FsmOptions): typeof Fsm;

    initialState: string;
    eventListeners: EventListeners;
    states: States;
    inputQueue: any[];
    namespace?: string;
    targetReplayState: any;
    state: string;
    priorState: string;
    priorAction: string;
    currentAction: string;
    currentActionArgs: any[];
    initialize: () => void;

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
}
