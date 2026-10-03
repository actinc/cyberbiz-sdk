<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz;

use Actinc\Cyberbiz\Exception\DecodeException;

/**
 * Typed, path-aware reads from one decoded JSON object, for building models.
 * Every failure names the field, e.g. "$.line_items[2].price: expected
 * money, got array", so a Golden File mismatch points at the exact field.
 */
final class Fields
{
    /** @param array<array-key, mixed> $data */
    private function __construct(private readonly array $data, public readonly string $path) {}

    /** @throws DecodeException when $value is not a JSON object */
    public static function of(mixed $value, string $path = '$'): self
    {
        if (!\is_array($value) || ($value !== [] && array_is_list($value))) {
            throw self::mismatch($path, 'object', $value);
        }

        return new self($value, $path);
    }

    public function has(string $key): bool
    {
        return \array_key_exists($key, $this->data);
    }

    /** The raw decoded value, or null when absent. */
    public function raw(string $key): mixed
    {
        return $this->data[$key] ?? null;
    }

    public function int(string $key): int
    {
        return $this->intOrNull($key) ?? throw self::mismatch($this->at($key), 'int', null);
    }

    public function intOrNull(string $key): ?int
    {
        $value = $this->raw($key);
        if ($value === null || \is_int($value)) {
            return $value;
        }
        if (\is_string($value) && preg_match('/^-?\d{1,18}$/', $value) === 1) {
            return (int) $value;
        }

        throw self::mismatch($this->at($key), 'int', $value);
    }

    public function string(string $key): string
    {
        return $this->stringOrNull($key) ?? throw self::mismatch($this->at($key), 'string', null);
    }

    public function stringOrNull(string $key): ?string
    {
        $value = $this->raw($key);
        if ($value === null || \is_string($value)) {
            return $value;
        }

        throw self::mismatch($this->at($key), 'string', $value);
    }

    public function bool(string $key): bool
    {
        return $this->boolOrNull($key) ?? throw self::mismatch($this->at($key), 'bool', null);
    }

    public function boolOrNull(string $key): ?bool
    {
        $value = $this->raw($key);
        if ($value === null || \is_bool($value)) {
            return $value;
        }

        throw self::mismatch($this->at($key), 'bool', $value);
    }

    public function money(string $key): Money
    {
        return $this->moneyOrNull($key) ?? throw self::mismatch($this->at($key), 'money', null);
    }

    public function moneyOrNull(string $key): ?Money
    {
        $value = $this->raw($key);
        if ($value === null || $value === '') {
            return null;
        }
        if (!\is_int($value) && !\is_string($value)) {
            throw self::mismatch($this->at($key), 'money', $value);
        }
        try {
            return Money::of($value);
        } catch (DecodeException $e) {
            throw new DecodeException($this->at($key) . ': ' . $e->getMessage(), 0, $e);
        }
    }

    /** A CYBERBIZ timestamp in Asia/Taipei; null when absent, null or "". */
    public function time(string $key): ?\DateTimeImmutable
    {
        $value = $this->raw($key);
        if ($value !== null && !\is_string($value)) {
            throw self::mismatch($this->at($key), 'timestamp', $value);
        }
        try {
            return Time::parse($value);
        } catch (DecodeException $e) {
            throw new DecodeException($this->at($key) . ': ' . $e->getMessage(), 0, $e);
        }
    }

    public function object(string $key): self
    {
        return self::of($this->raw($key), $this->at($key));
    }

    public function objectOrNull(string $key): ?self
    {
        return $this->raw($key) === null ? null : $this->object($key);
    }

    /**
     * Maps a JSON array of objects; absent or null is an empty list.
     *
     * @template T
     *
     * @param callable(self): T $map
     *
     * @return list<T>
     */
    public function list(string $key, callable $map): array
    {
        $value = $this->raw($key) ?? [];
        if (!\is_array($value) || !array_is_list($value)) {
            throw self::mismatch($this->at($key), 'array', $value);
        }
        $out = [];
        foreach ($value as $i => $item) {
            $out[] = $map(self::of($item, $this->at($key) . '[' . $i . ']'));
        }

        return $out;
    }

    private function at(string $key): string
    {
        return $this->path . '.' . $key;
    }

    private static function mismatch(string $path, string $expected, mixed $value): DecodeException
    {
        return new DecodeException(\sprintf('%s: expected %s, got %s', $path, $expected, get_debug_type($value)));
    }
}
