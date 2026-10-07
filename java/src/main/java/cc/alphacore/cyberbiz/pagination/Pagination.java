package cc.alphacore.cyberbiz.pagination;

import cc.alphacore.cyberbiz.Response;

/**
 * What CYBERBIZ reports in the {@code X-Page}, {@code X-Per-Page}, {@code X-Offset}, {@code
 * X-Total}, {@code X-Total-Pages}, {@code X-Next-Page} and {@code X-Prev-Page} headers. A missing
 * or empty header reads as 0, so {@code nextPage} and {@code prevPage} are 0 when there is no such
 * page.
 *
 * @param page the page number, from 1
 * @param perPage items per page
 * @param offset items skipped before this page
 * @param total items across all pages
 * @param totalPages number of pages
 * @param nextPage the next page number, or 0 on the last page
 * @param prevPage the previous page number, or 0 on the first page
 */
public record Pagination(
    int page, int perPage, int offset, int total, int totalPages, int nextPage, int prevPage) {

  /** The largest page size the platform accepts, and the SDK's default when walking pages. */
  public static final int MAX_PER_PAGE = 50;

  /**
   * Reads the pagination headers of a response.
   *
   * @param response the response of a list endpoint
   * @return the pagination; all zeros when the headers are absent
   */
  public static Pagination from(Response response) {
    return new Pagination(
        header(response, "X-Page"),
        header(response, "X-Per-Page"),
        header(response, "X-Offset"),
        header(response, "X-Total"),
        header(response, "X-Total-Pages"),
        header(response, "X-Next-Page"),
        header(response, "X-Prev-Page"));
  }

  /** Whether the server reported a next page. */
  public boolean hasNext() {
    return nextPage > 0;
  }

  /** An absent, empty or non-numeric header is 0, as the PHP SDK reads it. */
  private static int header(Response response, String name) {
    String value = response.header(name).strip();
    return value.matches("\\d{1,9}") ? Integer.parseInt(value) : 0;
  }
}
