<?php

declare(strict_types=1);

$finder = (new PhpCsFixer\Finder())
    ->in([__DIR__ . '/src', __DIR__ . '/tests']);

return (new PhpCsFixer\Config())
    ->setRiskyAllowed(true)
    ->setCacheFile(__DIR__ . '/../.php-cs-fixer.cache')
    ->setRules([
        '@PER-CS2.0' => true,
        '@PHP82Migration' => true,
        'declare_strict_types' => true,
        'no_unused_imports' => true,
        'ordered_imports' => true,
        'strict_comparison' => true,
    ])
    ->setFinder($finder);
