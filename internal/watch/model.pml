/**
 * The Hypcast watch implementation uses a boolean "running" variable to track
 * whether a handler goroutine is prepared to handle new values. Writers start a
 * goroutine if one is not running, and running goroutines handle new values as
 * long as possible before exiting, at which point they set "running" to false.
 *
 * This [Promela] model is intended to formally verify several properties of
 * this approach, by exhaustively considering every possible interleaving of the
 * atomic statements and blocks in the model:
 *
 *   - A handler executes at least once after a write is performed.
 *   - A handler does not execute for the same value more than once.
 *   - A handler does not execute more than once concurrently.
 *
 * The model considers the case of one watcher in the presence of concurrent
 * writes, and does not attempt to model registration or cancellation of
 * additional handlers.
 *
 * Note that the purpose of the model is NOT to verify the implementation of
 * watches in Hypcast, as the unit tests do. It is to validate the _concept_ of
 * triggering handlers on the basis of a synchronized "running" variable, to
 * ensure the above properties hold if the Go implementation faithfully reflects
 * the model.
 *
 * See model_test.go for a harness that runs [Spin] to validate this model.
 *
 * [Promela]: https://en.wikipedia.org/wiki/Promela
 * [Spin]: https://spinroot.com/
 */


/* These correspond directly to fields of the watch struct in Hypcast. w_next
 * encodes both the next and ok fields of the struct, such that (next == 0)
 * represents (ok == false). */
byte w_next = 0;
bool w_running = false;

/* This corresponds to the watch.update method that writers invoke to push new
 * values into a watch. */
proctype update(byte val) {
  bool start = false;
  atomic { /* Performed under w.mu. */
    start = !w_running;
    w_next = val;
    w_running = true;
  }
  atomic { /* Performed outside w.mu. */
    if
    :: start -> run handler()
    :: else -> skip
    fi
  }
}

/* State for the handler to validate the properties expressed above. */
byte g_handling = 0;
byte g_sum = 0;

proctype handler() {
  byte next;
  bool stop;
  do
  ::
    atomic { /* Performed under w.mu. */
      next = w_next;
      stop = (next == 0);
      w_next = 0;
      w_running = !stop
    }
    if /* Performed outside w.mu. */
    :: (stop == true) -> break
    :: else ->
      g_handling = g_handling + 1;
      atomic {
        g_sum = g_sum + next;
        assert(g_sum == 2 || g_sum == 3 || g_sum == 5);
        assert(g_handling == 1);
        g_handling = 0
      }
    fi
  od
}

init {
  run update(2);
  run update(3);
  /* Enforce the handler executing at least once. If g_sum is never incremented,
   * init will finish in an invalid end state. */
  (g_sum > 0)
}
