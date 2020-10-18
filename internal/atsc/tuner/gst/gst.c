#include "gst.h"

const char* const HYPCAST_SINK_NAME_RAW = "raw";
const char* const HYPCAST_SINK_NAME_VIDEO = "video";
const char* const HYPCAST_SINK_NAME_AUDIO = "audio";

void hypcast_define_sink(GstElement* pipeline, char* sink_name, HypcastSinkRef* sink_ref) {
  GstElement* element = gst_bin_get_by_name(GST_BIN(pipeline), sink_name);
  g_object_set(element, "emit-signals", TRUE, NULL);
  g_signal_connect(element, "new-sample", G_CALLBACK(hypcast_sink_sample), sink_ref);
}

GstFlowReturn hypcast_sink_sample(GstElement* object, gpointer user_data) {
  GstSample* sample = NULL;
  g_signal_emit_by_name(object, "pull-sample", &sample);
  if (sample == NULL) {
    return GST_FLOW_OK;
  }

  GstBuffer* buffer = NULL;
  buffer = gst_sample_get_buffer(sample);
  if (buffer == NULL) {
    return GST_FLOW_OK;
  }

  gpointer copy = NULL;
  gsize copy_size = 0;
  gst_buffer_extract_dup(buffer, 0, gst_buffer_get_size(buffer), &copy, &copy_size);

  HypcastSinkRef* sink_ref = (HypcastSinkRef*) user_data;
  hypcastGlobalSink(sink_ref, copy, copy_size, GST_BUFFER_DURATION(buffer));

  free(copy);
  gst_sample_unref(sample);
  return GST_FLOW_OK;
}
