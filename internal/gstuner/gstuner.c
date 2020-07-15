#include "gstuner.h"

void hyp_gst_init(void) {
  gst_init(NULL, NULL);
  fprintf(stderr, "it works!\n");
}
