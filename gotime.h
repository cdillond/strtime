#include <stdint.h>

struct result {
  int64_t value;
  int64_t status;
};

struct result parse_time(const char *value, const char *format);

int64_t fmt_time(int64_t t, char buf[], int64_t max_len, const char *format);
