#ifndef HYP_GST_H
#define HYP_GST_H

#include <stdlib.h>
#include <stdio.h>

#include <glib.h>
#include <gst/gst.h>

extern const char* const HYP_SINK_NAME_RAW;
extern const char* const HYP_SINK_NAME_VIDEO;
extern const char* const HYP_SINK_NAME_AUDIO;

typedef struct HypSinkRef {
  unsigned int global_pipeline_id;
  unsigned int sink_type;
} HypSinkRef;

extern void hypGlobalSink(HypSinkRef *, void *, int, int);

void hyp_define_sink(GstElement *, char *, HypSinkRef *);
GstFlowReturn hyp_sink_sample(GstElement *, gpointer);

#endif
