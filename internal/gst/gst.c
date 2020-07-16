#include "gst.h"

GstFlowReturn hyp_receive_sample(GstElement *object, gpointer user_data) {
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

  int *bin_id_data = (int *) user_data;
  int bin_id = *bin_id_data;

  hypGoReceiveSample(bin_id, copy, copy_size, GST_BUFFER_DURATION(buffer));

  free(copy);
  gst_sample_unref(sample);
  return GST_FLOW_OK;
}

void hyp_define_receiver(GstElement *pipeline, char *bin_name, int bin_id) {
  GstElement *element = gst_bin_get_by_name(GST_BIN(pipeline), bin_name);

  int *bin_id_data = malloc(sizeof(int));
  *bin_id_data = bin_id;

  g_object_set(element, "emit-signals", TRUE, NULL);
  g_signal_connect(element, "new-sample", G_CALLBACK(hyp_receive_sample), bin_id_data);
}
