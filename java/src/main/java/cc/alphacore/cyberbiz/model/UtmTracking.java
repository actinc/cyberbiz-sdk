package cc.alphacore.cyberbiz.model;

import java.time.OffsetDateTime;

/**
 * The UTM attribution captured at checkout (custom feature).
 *
 * @param utmSource utm_source
 * @param utmMedium utm_medium
 * @param utmCampaign utm_campaign
 * @param utmContent utm_content
 * @param utmTerm utm_term
 * @param utmClickTime when the link was clicked
 */
public record UtmTracking(
    String utmSource,
    String utmMedium,
    String utmCampaign,
    String utmContent,
    String utmTerm,
    OffsetDateTime utmClickTime) {}
