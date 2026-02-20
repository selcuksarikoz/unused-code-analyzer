<?php
// main.php
require_once 'utils.php';

use Utils\UsedClass;
use function Utils\used_function;
use const Utils\USED_CONST;

$obj = new UsedClass();
echo $obj->hello();

echo used_function();
echo USED_CONST;
