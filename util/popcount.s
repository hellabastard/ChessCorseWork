// popcount.s
#include "textflag.h"

// func PopCount(x uint64) int
TEXT ·PopCount(SB), NOSPLIT, $0-16
    MOVQ x+0(FP), AX    // Загружаем аргумент x (uint64) в регистр AX
    POPCNTQ AX, AX      // Используем инструкцию POPCNT для подсчета битов
    MOVQ AX, ret+8(FP)  // Сохраняем результат в возвращаемое значение
    RET
