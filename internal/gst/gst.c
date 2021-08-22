#include "gst.h"

void hypcast_connect_sink(GstElement *element, uintptr_t sink_handle) {
  g_object_set(element, "emit-signals", TRUE, NULL);
  g_signal_connect(element, "new-sample", G_CALLBACK(hypcast_sink_sample),
                   GSIZE_TO_POINTER(sink_handle));
}

GstFlowReturn hypcast_sink_sample(GstElement *object, gpointer user_data) {
  uintptr_t sink_handle = GPOINTER_TO_SIZE(user_data);

  GstSample *sample = NULL;
  g_signal_emit_by_name(object, "pull-sample", &sample);
  if (sample == NULL) {
    return GST_FLOW_OK;
  }

  // At this point, the Go side takes over the ownership of sample.
  return hypcastSinkSample(sample, sink_handle);
}
