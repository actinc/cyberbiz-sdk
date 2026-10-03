<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Webhook;

/**
 * HMAC-SHA256 webhook signatures. CYBERBIZ's documentation says the digest is
 * base64, but every production delivery carries 64 lowercase hex
 * characters, so verification accepts hex and base64, both in constant time.
 */
final class Signature
{
    /** The Signature CYBERBIZ sends for $body: lowercase hex HMAC-SHA256 keyed by the App Secret. */
    public static function sign(string $body, string $secret): string
    {
        return hash_hmac('sha256', $body, $secret);
    }

    /** The Domain Signature for a Shop Domain. */
    public static function signDomain(string $shopDomain, string $secret): string
    {
        return hash_hmac('sha256', $shopDomain, $secret);
    }

    /** Whether $signature (hex, either case, or base64) is the HMAC of $body; empty never verifies. */
    public static function verify(string $body, string $signature, string $secret): bool
    {
        return self::matches(hash_hmac('sha256', $body, $secret, true), $signature, $secret);
    }

    public static function verifyDomain(string $shopDomain, string $signature, string $secret): bool
    {
        return self::matches(hash_hmac('sha256', $shopDomain, $secret, true), $signature, $secret);
    }

    /** Both encodings are always compared, so timing does not reveal which matched. */
    private static function matches(string $digest, string $signature, string $secret): bool
    {
        $signature = trim($signature);
        if ($secret === '' || $signature === '') {
            return false;
        }
        $hexOk = hash_equals(bin2hex($digest), strtolower($signature));
        $base64Ok = hash_equals(base64_encode($digest), $signature);

        return $hexOk || $base64Ok;
    }
}
