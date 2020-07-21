#ifndef HYP_GSTUNER_H
#define HYP_GSTUNER_H

#include <stdlib.h>
#include <stdio.h>

#include <glib.h>
#include <gst/gst.h>

extern void hypGoSinkSample(int, void *, int, int);

void hyp_define_sink(GstElement *, char *, int);
GstFlowReturn hyp_sink_sample(GstElement *, gpointer);

#endif
