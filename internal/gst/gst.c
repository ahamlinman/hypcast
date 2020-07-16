#include "gst.h"

void hyp_define_sink(GstElement *pipeline, char *sink_name, int sink_type) {
  GstElement *element = gst_bin_get_by_name(GST_BIN(pipeline), sink_name);

  int *sink_type_data = malloc(sizeof(int));
  *sink_type_data = sink_type;

  g_object_set(element, "emit-signals", TRUE, NULL);
  g_signal_connect(element, "new-sample", G_CALLBACK(hyp_sink_sample), sink_type_data);
}

GstFlowReturn hyp_sink_sample(GstElement *object, gpointer user_data) {
  GstSample *sample = NULL;
  g_signal_emit_by_name(object, "pull-sample", &sample);
  if (sample == NULL) {
    return GST_FLOW_OK;
  }

  GstBuffer *buffer = NULL;
  buffer = gst_sample_get_buffer(sample);
  if (buffer == NULL) {
    return GST_FLOW_OK;
  }

  gpointer copy = NULL;
  gsize copy_size = 0;
  gst_buffer_extract_dup(buffer, 0, gst_buffer_get_size(buffer), &copy, &copy_size);

  int sink_type = * (int *) user_data;

  hypGoSinkSample(sink_type, copy, copy_size, GST_BUFFER_DURATION(buffer));

  free(copy);
  gst_sample_unref(sample);
  return GST_FLOW_OK;
}
