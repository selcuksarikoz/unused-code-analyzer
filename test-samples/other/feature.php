<?php
// feature.php
require_once 'utils.php';

use Utils\UsedClass;
use function Utils\used_function;
use function Utils\unused_function;
use const Utils\UNUSED_CONST;

class Feature {
    private $value;
    private $unusedValue;
    
    public function __construct() {
        $this->value = used_function();
    }
    
    public function getValue() {
        return $this->value;
    }
    
    private function unusedMethod() {
        return "unused";
    }
}

function usedFeatureFunction($used) {
    echo $used;
}

function unusedFeatureFunction($unused1, $unused2) {
    return "unused";
}

const USED_CONST = "used";
const UNUSED_CONST_FEATURE = "unused";

$obj = new UsedClass();
echo $obj->hello();

$unused_local = "unused";

function local_function() {
}
