// TODO: Declare stricter types. Consider contributing to machina.js itself [0]
// or DefinitelyTyped [1] (but preferably the former).
//
// [0]: https://github.com/ifandelse/machina.js/issues/158
// [1]: https://github.com/DefinitelyTyped/DefinitelyTyped/issues/25075
declare module 'machina' {
  // export as namespace machina;

  export class Fsm {
    new(...a: any[]): Fsm;
    static extend(...a: any[]): typeof Fsm;

    initialState: string;
    states: any;
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
