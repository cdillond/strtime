#define _XOPEN_SOURCE
#include "gotime.h"
#include <stdint.h>
#include <time.h>

#define ERR_PARSE 1
#define ERR_INVALID 2

/* returns the unix time (in seconds). */
struct result parse_time(const char *value, const char *format) {
  struct tm t = {};
  struct result res = {};

  if (!strptime(value, format, &t)) {
    res.status = ERR_PARSE;
    return res;
  }

  res.value = mktime(&t);
  if (res.value == -1) {
    res.status = ERR_INVALID;
  }

  return res;
}

int64_t fmt_time(int64_t t, char buf[], int64_t max_len, const char *format) {
  struct tm tm = {};
  time_t tp = t;
  if (!gmtime_r(&tp, &tm))
    return -ERR_INVALID;

  return strftime(buf, max_len, format, &tm);
}
