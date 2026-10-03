<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests;

use Actinc\Cyberbiz\Exception\DecodeException;
use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Json;
use Actinc\Cyberbiz\Response;

/**
 * Reads the shared Golden Files in testdata/golden/ (never copied into php/)
 * and decodes them the way the client does. A decoding failure names the
 * file as well as the field.
 */
final class Golden
{
    public const ROOT = __DIR__ . '/../../testdata/golden';

    /** The recorded response: body plus the headers saved next to it. */
    public static function response(string $name): Response
    {
        $body = self::read(self::ROOT . '/' . $name);
        $headersFile = self::ROOT . '/' . preg_replace('/\.json$/', '.headers.json', $name);
        $headers = [];
        if (is_file($headersFile)) {
            $decoded = json_decode(self::read($headersFile), true, 512, \JSON_THROW_ON_ERROR);
            foreach (\is_array($decoded) ? $decoded : [] as $key => $values) {
                $headers[(string) $key] = array_values(array_map(static fn(mixed $v): string => \is_scalar($v) ? (string) $v : '', \is_array($values) ? $values : [$values]));
            }
        }

        return new Response(200, $headers, $body);
    }

    /**
     * Decodes a JSON object file and builds a model from it.
     *
     * @template T
     *
     * @param string               $file  absolute path, or relative to testdata/golden
     * @param callable(Fields): T  $build
     *
     * @return T
     *
     * @throws DecodeException naming the file and the field
     */
    public static function object(string $file, callable $build): mixed
    {
        $path = str_starts_with($file, '/') ? $file : self::ROOT . '/' . $file;
        try {
            return $build(Fields::of(Json::decode(self::read($path))));
        } catch (DecodeException $e) {
            throw new DecodeException(basename($path) . ': ' . $e->getMessage(), 0, $e);
        }
    }

    private static function read(string $path): string
    {
        $data = file_get_contents($path);
        if ($data === false) {
            throw new \RuntimeException('cannot read Golden File ' . $path);
        }

        return $data;
    }
}
