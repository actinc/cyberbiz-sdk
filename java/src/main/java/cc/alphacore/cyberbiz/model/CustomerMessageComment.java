package cc.alphacore.cyberbiz.model;

import java.time.OffsetDateTime;

/**
 * One message in a customer service thread.
 *
 * @param id the comment id, or null
 * @param role who replied
 * @param admin admin details when role is admin
 * @param content the message
 * @param createdAt when it was posted
 */
public record CustomerMessageComment(
    Long id, String role, String admin, String content, OffsetDateTime createdAt) {}
