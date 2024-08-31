byte w_next = 0;
bool w_running = false;

proctype writer(byte val) {
  atomic {
    w_next = val;
    if
    :: (w_running == false) -> w_running = true; run handler()
    :: else -> skip
    fi
  }
}

byte g_handling = 0;
byte g_sum = 0;

proctype handler() {
  byte next;
  bool stop;

  do
  ::
    atomic {
      next = w_next;
      stop = (next == 0);
      w_next = 0;
      w_running = !stop
    }
    if
    :: (stop == true) -> break
    :: else ->
      g_handling = g_handling + 1;
      atomic {
        g_sum = g_sum + next;
        assert(g_sum == 11 || g_sum == 12 || g_sum == 23);
        assert(g_handling == 1);
        g_handling = 0
      }
    fi
  od
}

init {
  run writer(11);
  run writer(12);
  (g_sum > 0)
}
