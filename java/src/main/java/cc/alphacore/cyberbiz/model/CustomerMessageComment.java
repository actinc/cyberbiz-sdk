package cc.alphacore.cyberbiz.model;

import java.time.OffsetDateTime;

/**
 * One message in a customer service thread.
 *
 * @param id the comment id, or null
 * @param role who replied
 * @param admin the replying admin when role is admin, otherwise null
 * @param content the message
 * @param createdAt when it was posted
 */
public record CustomerMessageComment(
    Long id, String role, CustomerMessageAdmin admin, String content, OffsetDateTime createdAt) {}
