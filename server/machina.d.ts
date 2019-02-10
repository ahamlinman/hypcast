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

  export interface FsmOptions {
    initialState?: string;
    states?: States;
    eventListeners?: any[];
    namespace?: string;
    initialize?: () => void;
  }

  export class Fsm {
    new(options: FsmOptions): Fsm;
    static extend(options: FsmOptions): typeof Fsm;

    initialState: string;
    states: States;
    initialize: () => void;
    state: string;

    emit(event: string, ...args: any[]): void;
    handle(event: string, ...args: any[]): void;
    transition(state: string): void;
    deferUntilTransition(state?: string): void;
    on(event: string, callback: any): any;
    off(event?: string, callback?: any): void;
  }
}
