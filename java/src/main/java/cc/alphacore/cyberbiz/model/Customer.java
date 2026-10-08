package cc.alphacore.cyberbiz.model;

import cc.alphacore.cyberbiz.Money;
import java.time.OffsetDateTime;
import java.util.List;

/**
 * A shop member (GET /v1/customers, /v1/customers/{id}, /v2/customers, and embedded in an order).
 * The v1 detail response omits id; {@code CustomerService.get} fills it in. vipInfo is only present
 * on v2 lists requested with include_params=vip_info.
 *
 * @param id the customer id
 * @param name name
 * @param status "pending", "validate", "enabled", "disabled", "invited", "declined" or "warning"
 * @param email email
 * @param countryCallingCode e.g. "+886"
 * @param mobile mobile number
 * @param enableCvsPickup may pick up at convenience stores
 * @param enableCvsCod may pay cash on CVS pickup
 * @param enableHomeDeliveryCod may pay cash on home delivery, or null
 * @param acceptsMarketing accepts marketing
 * @param acceptsEmailNotification accepts email notifications
 * @param tags the customer's tags
 * @param address default address, or null
 * @param gender free text; shops define their own options
 * @param birthday a date (midnight, Asia/Taipei), or null
 * @param otherAccumulatedConsumption spend accumulated through other channels
 * @param otherAccumulatedConsumptionExpiredAt when that spend expires, or null
 * @param note staff note
 * @param customFields shop-defined fields
 * @param createdAt registered at
 * @param updatedAt last changed at
 * @param confirmedAt email verified at; null when never verified
 * @param mobileSmsConfirmedAt mobile verified at; null when never verified
 * @param bonusRemain remaining bonus points
 * @param uidProviders linked external login identities
 * @param vipInfo VIP state, or null
 */
public record Customer(
    long id,
    String name,
    String status,
    String email,
    String countryCallingCode,
    String mobile,
    boolean enableCvsPickup,
    boolean enableCvsCod,
    Boolean enableHomeDeliveryCod,
    boolean acceptsMarketing,
    boolean acceptsEmailNotification,
    List<Tag> tags,
    Address address,
    String gender,
    OffsetDateTime birthday,
    Money otherAccumulatedConsumption,
    OffsetDateTime otherAccumulatedConsumptionExpiredAt,
    String note,
    List<CustomField> customFields,
    OffsetDateTime createdAt,
    OffsetDateTime updatedAt,
    OffsetDateTime confirmedAt,
    OffsetDateTime mobileSmsConfirmedAt,
    Money bonusRemain,
    List<CustomerUidProvider> uidProviders,
    CustomerVipInfo vipInfo) {

  /** Copies the lists; a missing list becomes empty. */
  public Customer {
    tags = Lists.copy(tags);
    customFields = Lists.copy(customFields);
    uidProviders = Lists.copy(uidProviders);
  }

  /**
   * Returns a copy with another id; the v1 detail response omits it.
   *
   * @param newId the customer id
   * @return the copy
   */
  public Customer withId(long newId) {
    return new Customer(
        newId,
        name,
        status,
        email,
        countryCallingCode,
        mobile,
        enableCvsPickup,
        enableCvsCod,
        enableHomeDeliveryCod,
        acceptsMarketing,
        acceptsEmailNotification,
        tags,
        address,
        gender,
        birthday,
        otherAccumulatedConsumption,
        otherAccumulatedConsumptionExpiredAt,
        note,
        customFields,
        createdAt,
        updatedAt,
        confirmedAt,
        mobileSmsConfirmedAt,
        bonusRemain,
        uidProviders,
        vipInfo);
  }
}
