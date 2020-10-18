#ifndef HYPCAST_GST_H
#define HYPCAST_GST_H

#include <stdlib.h>
#include <stdio.h>

#include <glib.h>
#include <gst/gst.h>

extern const char* const HYPCAST_SINK_NAME_RAW;
extern const char* const HYPCAST_SINK_NAME_VIDEO;
extern const char* const HYPCAST_SINK_NAME_AUDIO;

typedef struct HypcastSinkRef {
  unsigned int global_pipeline_id;
  unsigned int sink_type;
} HypcastSinkRef;

extern void hypcastGlobalSink(HypcastSinkRef*, void*, int, int);

void hypcast_define_sink(GstElement*, char*, HypcastSinkRef*);
GstFlowReturn hypcast_sink_sample(GstElement*, gpointer);

#endif
