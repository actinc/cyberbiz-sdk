package cc.alphacore.cyberbiz.model;

/**
 * A customer's VIP membership state (GET /v1/customers/{id}/vip_info); group and levels are null
 * outside a VIP programme.
 *
 * @param customerId the customer id
 * @param currentGroup the group, or null
 * @param currentLevel the current level, or null
 * @param nextLevel the next level, or null
 * @param extraInfo validity and distance to renew or upgrade, or null
 */
public record CustomerVipInfo(
    long customerId,
    CustomerVipGroup currentGroup,
    CustomerVipLevel currentLevel,
    CustomerVipLevel nextLevel,
    CustomerVipExtraInfo extraInfo) {}
