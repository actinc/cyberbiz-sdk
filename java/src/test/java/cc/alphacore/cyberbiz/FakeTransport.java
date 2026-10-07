package cc.alphacore.cyberbiz;

import cc.alphacore.cyberbiz.http.Transport;
import cc.alphacore.cyberbiz.http.TransportRequest;
import java.io.IOException;
import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Deque;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/** Replies with a scripted sequence of responses or I/O failures and records every request. */
final class FakeTransport implements Transport {
  private final Deque<Object> replies = new ArrayDeque<>();
  final List<TransportRequest> requests = new ArrayList<>();

  FakeTransport reply(int status, String body, String... headerPairs) {
    Map<String, List<String>> headers = new LinkedHashMap<>();
    for (int i = 0; i + 1 < headerPairs.length; i += 2) {
      headers.put(headerPairs[i], List.of(headerPairs[i + 1]));
    }
    replies.add(new Response(status, headers, body));
    return this;
  }

  FakeTransport fail(String message) {
    replies.add(new IOException(message));
    return this;
  }

  @Override
  public synchronized Response send(TransportRequest request) throws IOException {
    requests.add(request);
    Object next = replies.poll();
    if (next == null) {
      throw new AssertionError("unexpected request " + request);
    }
    if (next instanceof IOException e) {
      throw e;
    }
    return (Response) next;
  }
}
