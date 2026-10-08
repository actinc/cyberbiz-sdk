package cc.alphacore.cyberbiz.model;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/** The immutable copies the model records make of their lists. */
final class Lists {
  private Lists() {}

  /**
   * Copies a decoded list into an unmodifiable one. A missing or null list becomes empty; null
   * elements are kept, since {@code List.copyOf} would reject what the API sent.
   *
   * @param list the list, or null
   * @param <T> the element type
   * @return an unmodifiable copy
   */
  static <T> List<T> copy(List<T> list) {
    return list == null ? List.of() : Collections.unmodifiableList(new ArrayList<>(list));
  }
}
