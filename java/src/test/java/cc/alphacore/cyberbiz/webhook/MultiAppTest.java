package cc.alphacore.cyberbiz.webhook;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import cc.alphacore.cyberbiz.exception.AmbiguousSecretException;
import cc.alphacore.cyberbiz.exception.InvalidDomainSignatureException;
import cc.alphacore.cyberbiz.exception.InvalidSignatureException;
import cc.alphacore.cyberbiz.exception.TooManyCredentialsException;
import cc.alphacore.cyberbiz.exception.UnknownShopException;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import org.junit.jupiter.api.Test;

/**
 * Several Apps on one Shop (CBSDK-41): CYBERBIZ sends no App identifier, so the App is the one
 * whose secret verifies the body Signature.
 */
class MultiAppTest {

  private static final String SHOP = "shop-a.cyberbiz.co";
  private static final String SECRET_A = "app-a-test-secret";
  private static final String SECRET_B = "app-b-test-secret";
  private static final byte[] BODY =
      "{\"id\":1001,\"name\":\"#1001\"}".getBytes(StandardCharsets.UTF_8);
  private static final WebhookParser TWO_APPS =
      new WebhookParser(new AppSecrets(Map.of(SHOP, Map.of("app-a", SECRET_A, "app-b", SECRET_B))));

  /** A delivery for SHOP signed with {@code secret}. */
  private static Map<String, String> signedWith(String secret) {
    Map<String, String> headers = new HashMap<>();
    headers.put(WebhookParser.HEADER_EVENT, "orders/paid");
    headers.put(WebhookParser.HEADER_DOMAIN, SHOP);
    headers.put(WebhookParser.HEADER_SIGNATURE, Signature.sign(BODY, secret));
    return headers;
  }

  /** {@code count} candidates "app-i" with secret "secret-i". */
  private static List<Credential> numbered(int count) {
    List<Credential> out = new ArrayList<>();
    for (int i = 0; i < count; i++) {
      out.add(new Credential("app-" + i, "secret-" + i));
    }
    return out;
  }

  @Test
  void identifiesTheAppWhoseSecretVerified() { // AC1
    assertEquals("app-b", TWO_APPS.parse(signedWith(SECRET_B), BODY).appId());
    assertEquals("app-a", TWO_APPS.parse(signedWith(SECRET_A), BODY).appId());
  }

  @Test
  void rejectsWhenNoCandidateVerifies() { // AC2
    InvalidSignatureException e =
        assertThrows(
            InvalidSignatureException.class,
            () -> TWO_APPS.parse(signedWith("app-c-test-secret"), BODY));
    assertEquals(401, e.httpStatus());
  }

  @Test
  void rejectsTwoAppsSharingASecret() { // AC3
    WebhookParser same =
        new WebhookParser(
            new AppSecrets(Map.of(SHOP, Map.of("app-a", SECRET_A, "app-b", SECRET_A))));
    AmbiguousSecretException e =
        assertThrows(AmbiguousSecretException.class, () -> same.parse(signedWith(SECRET_A), BODY));
    assertEquals(500, e.httpStatus());
  }

  @Test
  void rejectsMoreCandidatesThanTheCap() { // AC5
    List<Credential> many = numbered(WebhookParser.MAX_CREDENTIALS + 1);
    WebhookParser parser = new WebhookParser(SecretResolver.ofCredentials(shop -> many));
    TooManyCredentialsException e =
        assertThrows(
            TooManyCredentialsException.class, () -> parser.parse(signedWith("secret-0"), BODY));
    assertEquals(500, e.httpStatus());
  }

  @Test
  void acceptsExactlyTheCap() {
    List<Credential> atCap = numbered(WebhookParser.MAX_CREDENTIALS);
    WebhookParser parser = new WebhookParser(SecretResolver.ofCredentials(shop -> atCap));

    assertEquals("app-15", parser.parse(signedWith("secret-15"), BODY).appId());
  }

  @Test
  void verifiesTheDomainSignatureWithTheMatchedSecret() {
    Map<String, String> ok = signedWith(SECRET_B);
    ok.put(WebhookParser.HEADER_DOMAIN_SIGNATURE, Signature.signDomain(SHOP, SECRET_B));
    assertEquals("app-b", TWO_APPS.parse(ok, BODY).appId());

    Map<String, String> other = signedWith(SECRET_B);
    other.put(WebhookParser.HEADER_DOMAIN_SIGNATURE, Signature.signDomain(SHOP, SECRET_A));
    assertThrows(InvalidDomainSignatureException.class, () -> TWO_APPS.parse(other, BODY));
  }

  @Test
  void ignoresEmptySecrets() {
    WebhookParser parser =
        new WebhookParser(
            SecretResolver.ofCredentials(
                shop -> List.of(new Credential("app-a", ""), new Credential("app-b", SECRET_B))));

    assertEquals("app-b", parser.parse(signedWith(SECRET_B), BODY).appId());
  }

  @Test
  void failsClosedWithoutSecrets() {
    List<SecretResolver> resolvers =
        List.of(
            SecretResolver.ofCredentials(shop -> List.of()),
            SecretResolver.ofCredentials(shop -> null),
            SecretResolver.ofCredentials(shop -> Arrays.asList((Credential) null)),
            SecretResolver.ofCredentials(
                shop -> List.of(new Credential("app-a", ""), new Credential("app-b", ""))),
            new AppSecrets(Map.of("shop-b.cyberbiz.co", Map.of("app-a", SECRET_A))));
    for (SecretResolver resolver : resolvers) {
      WebhookParser parser = new WebhookParser(resolver);
      assertThrows(UnknownShopException.class, () -> parser.parse(signedWith(SECRET_A), BODY));
    }
  }

  @Test
  void singleSecretResolversLeaveTheAppIdEmpty() {
    List<SecretResolver> resolvers =
        List.of(
            new StaticSecret(SECRET_A),
            new SecretMap(Map.of(SHOP, SECRET_A)),
            shop -> Optional.of(SECRET_A));
    for (SecretResolver resolver : resolvers) {
      assertEquals("", new WebhookParser(resolver).parse(signedWith(SECRET_A), BODY).appId());
    }
  }

  @Test
  void appSecretsMatchesDomainsIgnoringCase() {
    WebhookParser parser =
        new WebhookParser(new AppSecrets(Map.of("Shop-A.cyberbiz.co", Map.of("app-b", SECRET_B))));

    assertEquals("app-b", parser.parse(signedWith(SECRET_B), BODY).appId());
  }

  @Test
  void secretForAnswersOnlyForASingleApp() {
    AppSecrets one = new AppSecrets(Map.of(SHOP, Map.of("app-a", SECRET_A)));
    assertEquals(Optional.of(SECRET_A), one.secretFor(SHOP));
    AppSecrets two = new AppSecrets(Map.of(SHOP, Map.of("app-a", SECRET_A, "app-b", SECRET_B)));
    assertEquals(Optional.empty(), two.secretFor(SHOP));
    SecretResolver lookup = SecretResolver.ofCredentials(shop -> List.of());
    assertEquals(Optional.empty(), lookup.secretFor(SHOP));
  }

  @Test
  void neverShowsTheSecrets() {
    String shown =
        new Credential("app-a", SECRET_A)
            + " "
            + new AppSecrets(Map.of(SHOP, Map.of("app-a", SECRET_A)));

    assertFalse(shown.contains(SECRET_A), shown);
    assertTrue(shown.contains("app-a"), shown);
  }

  @Test
  void eventKeepsItsSevenArgumentConstructor() {
    Event e = new Event("orders/paid", SHOP, "", "sig", "", Map.of(), "{}");

    assertEquals("", e.appId());
  }
}
