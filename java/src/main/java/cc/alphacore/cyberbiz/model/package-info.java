/**
 * The API's response models, as immutable records decoded by {@link
 * cc.alphacore.cyberbiz.json.Json}.
 *
 * <ul>
 *   <li>Amounts are {@link cc.alphacore.cyberbiz.Money} and times {@link java.time.OffsetDateTime}
 *       in Asia/Taipei; both are null when the API sends null or omits the field.
 *   <li>Strings, nested objects and amounts are null when absent; lists are never null (a missing
 *       list is empty) and unmodifiable.
 *   <li>A number or boolean is a primitive only where every Golden File carries a value; where the
 *       Golden Files show null, or never show the field, it is a box type ({@code Long}, {@code
 *       Integer}, {@code Boolean}, {@code Double}) that may be null. A primitive the response omits
 *       reads as 0 or false.
 * </ul>
 */
package cc.alphacore.cyberbiz.model;
