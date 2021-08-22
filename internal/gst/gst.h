#ifndef HYPCAST_GST_H
#define HYPCAST_GST_H

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

#include <glib.h>
#include <gst/gst.h>

extern GstFlowReturn hypcastSinkSample(GstSample *, uintptr_t);

void hypcast_connect_sink(GstElement *, uintptr_t);
GstFlowReturn hypcast_sink_sample(GstElement *, gpointer);

#endif
