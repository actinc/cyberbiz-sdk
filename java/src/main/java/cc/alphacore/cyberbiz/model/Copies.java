package cc.alphacore.cyberbiz.model;

import com.google.gson.JsonElement;
import com.google.gson.JsonNull;

/** The defensive copies the model records make of free-form JSON; lists use {@link Lists}. */
final class Copies {
  private Copies() {}

  /** A deep copy of free-form JSON; an absent value becomes {@link JsonNull#INSTANCE}. */
  static JsonElement json(JsonElement value) {
    return value == null ? JsonNull.INSTANCE : value.deepCopy();
  }
}
