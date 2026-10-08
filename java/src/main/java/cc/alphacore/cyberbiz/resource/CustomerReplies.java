package cc.alphacore.cyberbiz.resource;

/** The single-field replies {@link CustomerService} unwraps into plain values. */
final class CustomerReplies {
  private CustomerReplies() {}

  /**
   * One item of GET /v1/customers/default_gender_options.
   *
   * @param defaultGenderOption the option
   */
  record GenderOption(String defaultGenderOption) {}

  /**
   * The reply of GET /v1/customers/{id}/account_activation_url.
   *
   * @param accountActivationUrl the link
   */
  record ActivationUrl(String accountActivationUrl) {}
}
