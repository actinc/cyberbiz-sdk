package cc.alphacore.cyberbiz.pagination;

import java.util.Collections;
import java.util.Iterator;
import java.util.NoSuchElementException;
import java.util.Objects;
import java.util.Spliterator;
import java.util.Spliterators;
import java.util.function.IntFunction;
import java.util.stream.Stream;
import java.util.stream.StreamSupport;

/**
 * Every item of a list endpoint, fetched lazily one page at a time. Nothing is requested until
 * iteration starts; each page is requested only when the previous one is used up, following {@code
 * X-Next-Page}, and the walk stops at the last page or at an empty page, so it sends exactly one
 * request per page. Each call to {@link #iterator()} or {@link #stream()} starts a new walk from
 * the first page. An API error while walking is thrown from {@code hasNext()} / {@code next()}.
 *
 * <pre>{@code
 * for (Product product : client.all(Request.of("GET", "/v1/products"), Product.class)) { ... }
 * client.all(request, Product.class).stream().filter(...).toList();
 * }</pre>
 *
 * @param <T> the item type
 */
public final class PageIterable<T> implements Iterable<T> {
  private final IntFunction<Page<T>> fetch;
  private final int firstPage;

  /**
   * Creates the walk; normally obtained from {@code CyberbizClient.all}.
   *
   * @param fetch requests the page with the given number
   * @param firstPage the page to start at, from 1
   */
  public PageIterable(IntFunction<Page<T>> fetch, int firstPage) {
    this.fetch = Objects.requireNonNull(fetch, "fetch");
    if (firstPage < 1) {
      throw new IllegalArgumentException("firstPage must be at least 1: " + firstPage);
    }
    this.firstPage = firstPage;
  }

  /** Starts a new walk from the first page. */
  @Override
  public Iterator<T> iterator() {
    return new Walk();
  }

  /**
   * Streams the items; pages are requested as the stream pulls items, so a short-circuiting
   * operation such as {@code findFirst()} stops further requests.
   *
   * @return a sequential, ordered stream
   */
  public Stream<T> stream() {
    return StreamSupport.stream(
        Spliterators.spliteratorUnknownSize(iterator(), Spliterator.ORDERED), false);
  }

  /** The state of one walk: the items left on the current page and the page to fetch next. */
  private final class Walk implements Iterator<T> {
    private Iterator<T> current = Collections.emptyIterator();
    private int nextPage = firstPage;

    @Override
    public boolean hasNext() {
      while (!current.hasNext() && nextPage > 0) {
        Page<T> page = fetch.apply(nextPage);
        current = page.items().iterator();
        nextPage = page.hasNext() && !page.items().isEmpty() ? page.pagination().nextPage() : 0;
      }
      return current.hasNext();
    }

    @Override
    public T next() {
      if (!hasNext()) {
        throw new NoSuchElementException();
      }
      return current.next();
    }
  }
}
