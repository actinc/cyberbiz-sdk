<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests;

/**
 * The signed samples in docs/api/en/webhooks.md ("## Samples"), read in
 * place: one per event, with the HTTP headers, the base64 form of the
 * signature and the exact body (the JSON block plus its trailing newline),
 * all signed with the App Secret "example-app-secret".
 */
final class WebhookSamples
{
    public const SECRET = 'example-app-secret';

    public const FILE = __DIR__ . '/../../docs/api/en/webhooks.md';

    /** @return array<string, array{event: string, headers: array<string, string>, base64: string, body: string}> */
    public static function all(): array
    {
        $doc = file_get_contents(self::FILE);
        if ($doc === false) {
            throw new \RuntimeException('cannot read ' . self::FILE);
        }
        $samples = substr($doc, (int) strpos($doc, "\n## Samples"));
        preg_match_all('/^### `([^`]+)`\n.*?```http\n(.*?)```.*?`([A-Za-z0-9+\/=]{44})`.*?```json\n(.*?)```/ms', $samples, $matches, \PREG_SET_ORDER);

        $out = [];
        foreach ($matches as [, $event, $http, $base64, $json]) {
            $out[$event] = ['event' => $event, 'headers' => self::headers($http), 'base64' => $base64, 'body' => $json];
        }

        return $out;
    }

    /** @return array<string, string> */
    private static function headers(string $http): array
    {
        $headers = [];
        foreach (\array_slice(explode("\n", trim($http)), 1) as $line) {
            [$name, $value] = array_map(trim(...), explode(':', $line, 2)) + [1 => ''];
            $headers[$name] = $value;
        }

        return $headers;
    }
}
