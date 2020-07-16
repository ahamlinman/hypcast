#ifndef HYP_GSTUNER_H
#define HYP_GSTUNER_H

#include <stdlib.h>
#include <stdio.h>

#include <glib.h>
#include <gst/gst.h>

extern void hypGoReceiveSample(int, void *, int, int);

GstFlowReturn hyp_receive_sample(GstElement *, gpointer);
void hyp_define_receiver(GstElement *, char *, int);

#endif
