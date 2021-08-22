#ifndef HYPCAST_GST_H
#define HYPCAST_GST_H

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

#include <glib.h>
#include <gst/gst.h>

typedef struct HypcastSinkRef {
  uintptr_t handle;
  unsigned int index;
} HypcastSinkRef;

extern GstFlowReturn hypcastSinkSample(HypcastSinkRef *, GstSample *);

void hypcast_connect_sink(GstElement *, HypcastSinkRef *);
GstFlowReturn hypcast_sink_sample(GstElement *, gpointer);

#endif
