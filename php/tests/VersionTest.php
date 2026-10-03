<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests;

use Actinc\Cyberbiz\Client;
use PHPUnit\Framework\TestCase;

/** php/VERSION drives the release tag; the User-Agent must report the same version. */
final class VersionTest extends TestCase
{
    public function testClientVersionMatchesTheReleaseFile(): void
    {
        $file = file_get_contents(__DIR__ . '/../VERSION');
        self::assertNotFalse($file);
        self::assertSame(trim($file), Client::VERSION);
    }
}
