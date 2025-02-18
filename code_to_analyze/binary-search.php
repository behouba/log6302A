<?php
function binarySearch($arr, $target) {
    $low = 0;
    $high = count($arr) - 1;

    while ($low <= $high) {
        $mid = (int)(($low + $high) / 2);

        // Check if the target is present at mid
        if ($arr[$mid] == $target) {
            return $mid;
        }

        // If target is greater, ignore the left half
        if ($arr[$mid] < $target) {
            $low = $mid + 1;
        }
        // If target is smaller, ignore the right half
        else {
            $high = $mid - 1;
        }
    }

    // Target is not present in the array
    return -1;
}

// Example usage
$arr = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
$target = 7;

$result = binarySearch($arr, $target);

if ($result != -1) {
    echo "Element found at index " . $result;
} else {
    echo "Element not found in the array";
}
?>
