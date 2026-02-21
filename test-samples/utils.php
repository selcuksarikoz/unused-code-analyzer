<?php
// utils.php
namespace Utils;

class UsedClass {
    public function hello(): string {
        return "hello";
    }
}

const USED_CONST = "world";

function used_function(): string {
    return "function";
}

interface UsedInterface {
    public function getId(): int;
}

trait UsedTrait {
    public function traitMethod(): string {
        return "trait";
    }
}

function unused_function(): string {
    return "unused";
}

const UNUSED_CONST = 123;

class UnusedClass {
}
