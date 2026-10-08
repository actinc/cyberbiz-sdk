package cc.alphacore.cyberbiz.webhook;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Optional;

/**
 * A fixed Shop Domain to App ID to App Secret map, for one receiver that serves several Apps
 * installed on the same Shop. Domains match ignoring case.
 */
public final class AppSecrets implements SecretResolver {

  /** Candidates keyed by lower-cased Shop Domain. */
  private final Map<String, List<Credential>> credentials;

  /**
   * Creates the resolver from a copy of the map.
   *
   * @param secrets App Secrets keyed by Shop Domain, then by App ID, e.g. {@code
   *     Map.of("shop-a.cyberbiz.co", Map.of("app-a", secretA, "app-b", secretB))}
   */
  public AppSecrets(Map<String, Map<String, String>> secrets) {
    Map<String, List<Credential>> byShop = new HashMap<>();
    secrets.forEach(
        (domain, apps) -> {
          List<Credential> list =
              byShop.computeIfAbsent(domain.toLowerCase(Locale.ROOT), d -> new ArrayList<>());
          apps.forEach((appId, secret) -> list.add(new Credential(appId, secret)));
        });
    Map<String, List<Credential>> frozen = new HashMap<>();
    byShop.forEach((domain, list) -> frozen.put(domain, List.copyOf(list)));
    this.credentials = Map.copyOf(frozen);
  }

  @Override
  public List<Credential> credentialsFor(String shopDomain) {
    return credentials.getOrDefault(shopDomain.toLowerCase(Locale.ROOT), List.of());
  }

  /** Answers only when the Shop has exactly one App. */
  @Override
  public Optional<String> secretFor(String shopDomain) {
    List<Credential> list = credentialsFor(shopDomain);
    return list.size() == 1 && !list.get(0).secret().isEmpty()
        ? Optional.of(list.get(0).secret())
        : Optional.empty();
  }

  /** Lists the Shop Domains and App IDs, never the secrets. */
  @Override
  public String toString() {
    Map<String, List<String>> appIds = new HashMap<>();
    credentials.forEach(
        (domain, list) -> appIds.put(domain, list.stream().map(Credential::appId).toList()));
    return "AppSecrets" + appIds;
  }
}
