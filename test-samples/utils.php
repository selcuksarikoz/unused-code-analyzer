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

function unused_function(): string {
    return "unused";
}
