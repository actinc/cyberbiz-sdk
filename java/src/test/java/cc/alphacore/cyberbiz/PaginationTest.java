package cc.alphacore.cyberbiz;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import cc.alphacore.cyberbiz.exception.DecodeException;
import cc.alphacore.cyberbiz.exception.NotFoundException;
import cc.alphacore.cyberbiz.http.TransportRequest;
import cc.alphacore.cyberbiz.pagination.Page;
import cc.alphacore.cyberbiz.pagination.PageIterable;
import cc.alphacore.cyberbiz.pagination.Pagination;
import java.math.BigDecimal;
import java.net.URI;
import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;
import java.util.NoSuchElementException;
import java.util.stream.Collectors;
import java.util.stream.IntStream;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;

class PaginationTest {
  private static final Request PRODUCTS = Request.of("GET", "/v1/products");

  /** A synthetic list item. */
  record Item(long id, String title, BigDecimal price) {}

  private final FakeTransport transport = new FakeTransport();

  private CyberbizClient client() {
    return CyberbizClient.builder("synthetic-token-0123456789")
        .baseUrl(URI.create("https://api.example.test"))
        .transport(transport)
        .clock(new FakeClock())
        .rateLimit(0)
        .build();
  }

  /** Scripts page {@code page} of {@code pages} holding the given item ids. */
  private void page(int page, int pages, int... ids) {
    String body =
        IntStream.of(ids)
            .mapToObj(id -> "{\"id\":" + id + ",\"title\":\"Item " + id + "\",\"price\":0.1}")
            .collect(Collectors.joining(",", "[", "]"));
    transport.reply(
        200,
        body,
        "X-Page",
        String.valueOf(page),
        "X-Per-Page",
        "2",
        "X-Total-Pages",
        String.valueOf(pages),
        "X-Next-Page",
        page < pages ? String.valueOf(page + 1) : "",
        "X-Prev-Page",
        page > 1 ? String.valueOf(page - 1) : "");
  }

  private static List<Long> ids(Iterable<Item> items) {
    List<Long> ids = new ArrayList<>();
    items.forEach(item -> ids.add(item.id()));
    return ids;
  }

  private List<String> queries() {
    return transport.requests.stream().map(TransportRequest::uri).map(URI::getQuery).toList();
  }

  @Test
  void listReturnsOnePageWithItsPagination() {
    transport.reply(
        200,
        "[{\"id\":1,\"title\":\"Tea\",\"price\":\"120.50\",\"unknown_field\":[1]}]",
        "X-Page",
        "2",
        "X-Per-Page",
        "1",
        "X-Offset",
        "1",
        "X-Total",
        "3",
        "X-Total-Pages",
        "3",
        "X-Next-Page",
        "3",
        "X-Prev-Page",
        "1",
        "X-Request-Id",
        "REQ-1");

    Page<Item> page =
        client().list(PRODUCTS.withQuery("page", 2).withQuery("per_page", 1), Item.class);

    assertEquals(List.of(new Item(1, "Tea", new BigDecimal("120.50"))), page.items());
    assertEquals(new Pagination(2, 1, 1, 3, 3, 3, 1), page.pagination());
    assertTrue(page.hasNext());
    assertEquals("REQ-1", page.response().requestId());
    assertEquals(List.of("page=2&per_page=1"), queries());
    assertThrows(UnsupportedOperationException.class, () -> page.items().add(null));
  }

  @ParameterizedTest
  @ValueSource(strings = {"null", "", "  \n", "[]"})
  void listTreatsANullOrEmptyBodyAsNoItems(String body) {
    transport.reply(200, body);

    Page<Item> page = client().list(PRODUCTS, Item.class);

    assertEquals(List.of(), page.items());
    assertEquals(new Pagination(0, 0, 0, 0, 0, 0, 0), page.pagination());
    assertFalse(page.hasNext());
  }

  @Test
  void listRejectsABodyThatIsNotAnArray() {
    transport.reply(200, "{\"id\":1}");

    DecodeException e =
        assertThrows(DecodeException.class, () -> client().list(PRODUCTS, Item.class));

    assertEquals(
        "cyberbiz: GET /v1/products: expected a JSON array, got an object", e.getMessage());
  }

  @Test
  void listNamesTheRequestAndFieldOfAMistypedItem() {
    transport.reply(200, "[{\"id\":1},{\"id\":\"two\"}]");

    DecodeException e =
        assertThrows(DecodeException.class, () -> client().list(PRODUCTS, Item.class));

    assertTrue(e.getMessage().startsWith("cyberbiz: GET /v1/products: "), e.getMessage());
    assertTrue(e.getMessage().contains("$[1].id"), e.getMessage());
  }

  @Test
  void allWithZeroItemsSendsOneRequest() {
    page(1, 0);

    assertEquals(List.of(), ids(client().all(PRODUCTS, Item.class)));
    assertEquals(List.of("page=1&per_page=50"), queries());
  }

  @Test
  void allStopsAtAnEmptyPageEvenIfANextPageIsReported() {
    transport.reply(200, "[]", "X-Page", "1", "X-Next-Page", "2");

    assertEquals(List.of(), ids(client().all(PRODUCTS, Item.class)));
    assertEquals(1, transport.requests.size());
  }

  @Test
  void allWithExactlyOneFullPageSendsOneRequest() {
    page(1, 1, 1, 2);

    assertEquals(List.of(1L, 2L), ids(client().all(PRODUCTS.withQuery("per_page", 2), Item.class)));
    assertEquals(List.of("page=1&per_page=2"), queries());
  }

  @Test
  void allReadsAPartialLastPage() {
    page(1, 3, 1, 2);
    page(2, 3, 3, 4);
    page(3, 3, 5);

    assertEquals(
        List.of(1L, 2L, 3L, 4L, 5L),
        ids(client().all(PRODUCTS.withQuery("per_page", 2), Item.class)));
    assertEquals(List.of("page=1&per_page=2", "page=2&per_page=2", "page=3&per_page=2"), queries());
  }

  @ParameterizedTest
  @ValueSource(ints = {1, 2, 3, 7})
  void allSendsExactlyOneRequestPerPage(int pages) {
    for (int p = 1; p <= pages; p++) {
      page(p, pages, 2 * p - 1, 2 * p);
    }

    List<Item> items = client().all(PRODUCTS, Item.class).stream().toList();

    assertEquals(2 * pages, items.size());
    assertEquals(pages, transport.requests.size());
    assertEquals(new BigDecimal("0.1"), items.get(0).price());
  }

  @Test
  void allIsLazyAndStopsWhenTheStreamDoes() {
    page(1, 5, 1, 2);
    PageIterable<Item> all = client().all(PRODUCTS, Item.class);

    assertEquals(0, transport.requests.size());
    assertEquals(1L, all.stream().findFirst().orElseThrow().id());
    assertEquals(1, transport.requests.size());
  }

  @Test
  void allStartsAtTheRequestedPageAndKeepsTheOtherParameters() {
    page(3, 4, 5, 6);
    page(4, 4, 7);

    Request request = PRODUCTS.withQuery("status", "open").withQuery("page", 3);
    assertEquals(List.of(5L, 6L, 7L), ids(client().all(request, Item.class)));
    assertEquals(
        List.of("page=3&status=open&per_page=50", "page=4&status=open&per_page=50"), queries());
  }

  @Test
  void eachIterationStartsOverAndAnExhaustedIteratorThrows() {
    page(1, 1, 1);
    page(1, 1, 1);
    PageIterable<Item> all = client().all(PRODUCTS, Item.class);

    assertEquals(List.of(1L), ids(all));
    Iterator<Item> again = all.iterator();
    assertEquals(1L, again.next().id());
    assertFalse(again.hasNext());
    assertThrows(NoSuchElementException.class, again::next);
    assertEquals(2, transport.requests.size());
  }

  @Test
  void anApiErrorWhileWalkingIsThrown() {
    page(1, 2, 1, 2);
    transport.reply(404, "{\"error\":\"gone\"}");
    Iterator<Item> it = client().all(PRODUCTS, Item.class).iterator();

    it.next();
    it.next();
    assertThrows(NotFoundException.class, it::hasNext);
  }

  @ParameterizedTest
  @ValueSource(strings = {"0", "-2", "x", ""})
  void anInvalidStartPageMeansPageOne(String start) {
    assertEquals(1, PageRequests.startPage(PRODUCTS.withQuery("page", start)));
  }
}
