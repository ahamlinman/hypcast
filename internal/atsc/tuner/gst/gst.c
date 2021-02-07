#include "gst.h"

const char * const HYPCAST_SINK_NAME_RAW = "raw";
const char * const HYPCAST_SINK_NAME_VIDEO = "video";
const char * const HYPCAST_SINK_NAME_AUDIO = "audio";

void hypcast_define_sink(GstElement *pipeline, char *sink_name, HypcastSinkRef *sink_ref) {
  GstElement *element = gst_bin_get_by_name(GST_BIN(pipeline), sink_name);
  g_object_set(element, "emit-signals", TRUE, NULL);
  g_signal_connect(element, "new-sample", G_CALLBACK(hypcast_sink_sample), sink_ref);
}

GstFlowReturn hypcast_sink_sample(GstElement *object, gpointer user_data) {
  HypcastSinkRef *sink_ref = (HypcastSinkRef *) user_data;

  GstSample *sample = NULL;
  g_signal_emit_by_name(object, "pull-sample", &sample);
  if (sample == NULL) {
    return GST_FLOW_OK;
  }

  return hypcastSinkSample(sink_ref, sample); // Transfers ownership of sample
}
