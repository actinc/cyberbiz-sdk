package cc.alphacore.cyberbiz;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.io.IOException;
import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import org.junit.jupiter.api.Test;

class CyberbizClientTest {

  private static final String TOKEN = "synthetic-token-0123456789";

  @Test
  void buildsWithTheDefaultBaseUrl() {
    CyberbizClient client = CyberbizClient.builder(TOKEN).build();

    assertEquals(CyberbizClient.DEFAULT_BASE_URL, client.baseUrl());
  }

  @Test
  void overridesTheBaseUrlAndEndsItInASlash() {
    URI mock = URI.create("http://127.0.0.1:8080");

    assertEquals(
        URI.create("http://127.0.0.1:8080/"),
        CyberbizClient.builder(TOKEN).baseUrl(mock).build().baseUrl());
  }

  @Test
  void rejectsNegativeLimits() {
    CyberbizClient.Builder builder = CyberbizClient.builder(TOKEN);

    assertThrows(IllegalArgumentException.class, () -> builder.maxRetries(-1));
    assertThrows(IllegalArgumentException.class, () -> builder.rateLimit(-1));
  }

  @Test
  void rejectsABlankToken() {
    assertThrows(IllegalArgumentException.class, () -> CyberbizClient.builder(" "));
    assertThrows(NullPointerException.class, () -> CyberbizClient.builder(null));
  }

  @Test
  void rejectsANonHttpBaseUrl() {
    CyberbizClient.Builder builder = CyberbizClient.builder(TOKEN);

    assertThrows(IllegalArgumentException.class, () -> builder.baseUrl(URI.create("ftp://x/")));
    assertThrows(IllegalArgumentException.class, () -> builder.baseUrl(URI.create("/v1")));
  }

  @Test
  void toStringNeverShowsTheToken() {
    assertFalse(CyberbizClient.builder(TOKEN).build().toString().contains(TOKEN));
  }

  @Test
  void versionMatchesTheVersionFileAndTheArtifact() throws IOException {
    String file = Files.readString(Path.of("VERSION"), StandardCharsets.UTF_8).strip();

    assertEquals(file, CyberbizClient.VERSION, "CyberbizClient.VERSION vs java/VERSION");
    assertEquals(file, System.getProperty("project.version"), "pom.xml version vs java/VERSION");
  }
}
