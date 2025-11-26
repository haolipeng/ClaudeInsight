#ifndef __P_INFER_H__
#define __P_INFER_H__

#include "pktlatency.h"

static __always_inline int32_t read_big_endian_int32(const char* buf) {
  int32_t length;
  bpf_probe_read(&length, sizeof(length), buf);
  return bpf_ntohl(length);
}

static __always_inline int16_t read_big_endian_int16(const char* buf) {
  int16_t val;
  bpf_probe_read(&val, sizeof(val), buf);
  return bpf_ntohs(val);
}

static __always_inline enum message_type_t is_http_protocol(const char *old_buf, size_t count) {
  if (count < 5) {
    return 0;
  }
  char buf[4] = {};
  bpf_probe_read_user(buf, 4, old_buf);
  if (buf[0] == 'H' && buf[1] == 'T' && buf[2] == 'T' && buf[3] == 'P') {
    return kResponse;
  }
  if (buf[0] == 'G' && buf[1] == 'E' && buf[2] == 'T') {
    return kRequest;
  }
  if (buf[0] == 'H' && buf[1] == 'E' && buf[2] == 'A' && buf[3] == 'D') {
    return kRequest;
  }
  if (buf[0] == 'P' && buf[1] == 'O' && buf[2] == 'S' && buf[3] == 'T') {
    return kRequest;
  }
  return kUnknown;
}


static __inline enum message_type_t is_dns_protocol(const char* buf, size_t count) {
  const int kDNSHeaderSize = 12;

  // Use the maximum *guaranteed* UDP packet size as the max DNS message size.
  // UDP packets can be larger, but this is the typical maximum size for DNS.
  const int kMaxDNSMessageSize = 512;

  // Maximum number of resource records.
  // https://stackoverflow.com/questions/6794926/how-many-a-records-can-fit-in-a-single-dns-response
  const int kMaxNumRR = 25;

  if (count < kDNSHeaderSize || count > kMaxDNSMessageSize) {
    return kUnknown;
  }

  uint8_t ubuf[12] = {};
  bpf_probe_read_user(ubuf, 12, buf);

  uint16_t flags = (ubuf[2] << 8) + ubuf[3];
  uint16_t num_questions = (ubuf[4] << 8) + ubuf[5];
  uint16_t num_answers = (ubuf[6] << 8) + ubuf[7];
  uint16_t num_auth = (ubuf[8] << 8) + ubuf[9];
  uint16_t num_addl = (ubuf[10] << 8) + ubuf[11];

  bool qr = (flags >> 15) & 0x1;
  uint8_t opcode = (flags >> 11) & 0xf;
  uint8_t zero = (flags >> 6) & 0x1;

  if (zero != 0) {
    return kUnknown;
  }

  if (opcode != 0) {
    return kUnknown;
  }

  if (num_questions == 0 || num_questions > 10) {
    return kUnknown;
  }

  uint32_t num_rr = num_questions + num_answers + num_auth + num_addl;
  if (num_rr > kMaxNumRR) {
    return kUnknown;
  }
  return (qr == 0) ? kRequest : kResponse;
}

#define TRACE_PROTOCOL(p) (trace_protocol == kProtocolUnset || trace_protocol == p)

static __always_inline struct protocol_message_t infer_protocol(const char *buf, size_t count, 
    size_t total_count, struct conn_info_t *conn_info, enum traffic_protocol_t trace_protocol) {
  struct protocol_message_t protocol_message;
  protocol_message.protocol = kProtocolUnknown;
  protocol_message.type = kUnknown;
  conn_info->prepend_length_header = false;

  if (TRACE_PROTOCOL(kProtocolHTTP) && (protocol_message.type = is_http_protocol(buf, count)) != kUnknown) {
    protocol_message.protocol = kProtocolHTTP;
  } else if (TRACE_PROTOCOL(kProtocolDNS) && (protocol_message.type = is_dns_protocol(buf, count)) != kUnknown) {
    protocol_message.protocol = kProtocolDNS;
  }
  conn_info->prev_count = count;
  if (count == 4) {
    bpf_probe_read(conn_info->prev_buf, 4, buf);
  }
  return protocol_message;
}
#endif
