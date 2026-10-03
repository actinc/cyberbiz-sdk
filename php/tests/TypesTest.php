<?php

declare(strict_types=1);

namespace Actinc\Cyberbiz\Tests;

use Actinc\Cyberbiz\Exception\DecodeException;
use Actinc\Cyberbiz\Fields;
use Actinc\Cyberbiz\Json;
use Actinc\Cyberbiz\Money;
use Actinc\Cyberbiz\Time;
use PHPUnit\Framework\TestCase;

final class TypesTest extends TestCase
{
    public function testJsonKeepsDecimalsAsTheirExactText(): void
    {
        $decoded = Json::decode('{"price":3690.0,"qty":2,"big":12345678901234567890,"note":"1.5 \"in\" text","e":1e3,"list":[0.1,-2.50]}');

        self::assertSame([
            'price' => '3690.0',
            'qty' => 2,
            'big' => '12345678901234567890',
            'note' => '1.5 "in" text',
            'e' => '1e3',
            'list' => ['0.1', '-2.50'],
        ], $decoded);
    }

    public function testJsonRejectsMalformedInput(): void
    {
        $this->expectException(DecodeException::class);
        Json::decode('{"a":');
    }

    public function testMoneyIsExactWhereFloatIsNot(): void
    {
        self::assertNotSame(0.3, 0.1 + 0.2);
        self::assertTrue(Money::of('0.1')->add(Money::of('0.2'))->equals(Money::of('0.3')));
        self::assertSame('3690.00', Money::of('3690.0')->amount);
        self::assertSame('199.00', (string) Money::of(199));
        self::assertSame('12345678901234567.89', Money::of('12345678901234567.89')->amount);
        self::assertSame('1000.00', Money::of('1e3')->amount);
    }

    public function testMoneyRoundsHalfAwayFromZero(): void
    {
        self::assertSame('1.01', Money::of('1.005')->amount);
        self::assertSame('-1.01', Money::of('-1.005')->amount);
        self::assertSame('0.00', Money::of('-0.001')->amount);
        self::assertSame(-1, Money::of('9.99')->compare(Money::of('10')));
        self::assertTrue(Money::of('-0.5')->isNegative());
    }

    public function testMoneyRejectsNonNumbers(): void
    {
        $this->expectException(DecodeException::class);
        Money::of('12 TWD');
    }

    public function testTimesAreInTaipei(): void
    {
        $plain = Time::parse('2026-09-01 10:00:00');
        self::assertNotNull($plain);
        self::assertSame('Asia/Taipei', $plain->getTimezone()->getName());
        self::assertSame('2026-09-01 10:00:00', Time::format($plain));

        self::assertSame('2026-09-01 10:00:00', Time::format(Time::parse('2026-09-01T02:00:00Z') ?? self::fail()));
        self::assertSame('2026-09-01 10:00:00', Time::format(Time::parse('2026-09-01T10:00:00+08:00') ?? self::fail()));
        self::assertSame('2026-09-01 00:00:00', Time::format(Time::parse('2026-09-01') ?? self::fail()));
        self::assertSame('2026-09-01 10:00:00', Time::format(new \DateTimeImmutable('2026-09-01 02:00:00 UTC')));
        self::assertNull(Time::parse(null));
        self::assertNull(Time::parse(''));
    }

    public function testRejectsUnknownTimeFormats(): void
    {
        $this->expectException(DecodeException::class);
        Time::parse('next tuesday');
    }

    public function testFieldsReadTypedValues(): void
    {
        $f = Fields::of(Json::decode('{"id":"7","name":"範例","on":true,"price":200.0,"at":"2026-07-10 20:22:05","tags":[{"n":"a"}],"none":null}'));

        self::assertSame(7, $f->int('id'));
        self::assertSame('範例', $f->string('name'));
        self::assertTrue($f->bool('on'));
        self::assertSame('200.00', $f->money('price')->amount);
        self::assertSame('2026-07-10 20:22:05', Time::format($f->time('at') ?? self::fail()));
        self::assertSame(['a'], $f->list('tags', static fn(Fields $t): string => $t->string('n')));
        self::assertNull($f->intOrNull('none'));
        self::assertNull($f->moneyOrNull('missing'));
        self::assertSame([], $f->list('missing', static fn(Fields $t): string => 'x'));
    }

    public function testFieldErrorsNameThePath(): void
    {
        $f = Fields::of(Json::decode('{"line_items":[{"price":1},{"price":[1]}]}'));

        $this->expectExceptionMessage('$.line_items[1].price: expected money, got array');
        $f->list('line_items', static fn(Fields $item): mixed => $item->money('price'));
    }

    public function testFieldsRequireAnObject(): void
    {
        $this->expectExceptionMessage('$: expected object, got array');
        Fields::of([1, 2]);
    }
}
