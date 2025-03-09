#include <stdint.h>

uint16_t diff(const char *old, const char *new_, char **out_buf, uint32_t *out_buf_len);
uint16_t diff_apply(const char *dir, const char *buf, uint32_t buf_len);
