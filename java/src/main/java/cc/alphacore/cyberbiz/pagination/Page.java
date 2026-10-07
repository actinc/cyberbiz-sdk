package cc.alphacore.cyberbiz.pagination;

import cc.alphacore.cyberbiz.Response;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Objects;

/**
 * One page of a list response.
 *
 * @param items the decoded items in the order the API returned them; unmodifiable
 * @param pagination the pagination headers of the response
 * @param response the raw response, e.g. for {@link Response#requestId()}
 * @param <T> the item type
 */
public record Page<T>(List<T> items, Pagination pagination, Response response) {

  /** Copies the items into an unmodifiable list. */
  public Page {
    items = Collections.unmodifiableList(new ArrayList<>(Objects.requireNonNull(items, "items")));
    Objects.requireNonNull(pagination, "pagination");
    Objects.requireNonNull(response, "response");
  }

  /** Whether the server reported a next page. */
  public boolean hasNext() {
    return pagination.hasNext();
  }
}
