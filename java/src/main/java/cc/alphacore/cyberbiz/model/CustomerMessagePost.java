package cc.alphacore.cyberbiz.model;

import java.time.OffsetDateTime;
import java.util.List;

/**
 * One customer service thread (GET /v1/customers/{id}/message_posts).
 *
 * @param id the thread id, or null
 * @param title subject
 * @param category category
 * @param order the order the thread is about, as text
 * @param status "reply_yet" or "replied"
 * @param comments the messages
 * @param createdAt when it was opened
 */
public record CustomerMessagePost(
    Long id,
    String title,
    String category,
    String order,
    String status,
    List<CustomerMessageComment> comments,
    OffsetDateTime createdAt) {

  /** Copies the lists; a missing list becomes empty. */
  public CustomerMessagePost {
    comments = Lists.copy(comments);
  }
}
