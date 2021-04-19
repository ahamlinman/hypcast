#include "gst.h"

void hypcast_connect_sink(GstElement *element, HypcastSinkRef *ref) {
  g_object_set(element, "emit-signals", TRUE, NULL);
  g_signal_connect(element, "new-sample", G_CALLBACK(hypcast_sink_sample), ref);
}

GstFlowReturn hypcast_sink_sample(GstElement *object, gpointer user_data) {
  HypcastSinkRef *sink_ref = (HypcastSinkRef *)user_data;

  GstSample *sample = NULL;
  g_signal_emit_by_name(object, "pull-sample", &sample);
  if (sample == NULL) {
    return GST_FLOW_OK;
  }

  return hypcastSinkSample(sink_ref, sample); // Transfers ownership of sample
}
