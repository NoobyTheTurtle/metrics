// Package staticlint реализует multichecker для статического анализа кода.
//
// Multichecker включает в себя:
//   - Стандартные анализаторы из golang.org/x/tools/go/analysis/passes
//   - Все анализаторы класса SA из staticcheck.io
//   - Дополнительные анализаторы из staticcheck.io (ST, S1000)
//   - Публичные анализаторы: bodyclose, errcheck
//   - Собственный анализатор osexit для запрета прямого вызова os.Exit в main функциях
//
// Запуск:
//
//	go run cmd/staticlint/main.go ./...
//
// или после сборки:
//
//	./staticlint ./...
//
// Анализаторы:
//
// Стандартные анализаторы (golang.org/x/tools/go/analysis/passes):
//   - asmdecl: проверяет соответствие объявлений на Go и ассемблере
//   - assign: проверяет бесполезные присваивания
//   - atomic: проверяет правильность использования пакета sync/atomic
//   - bools: проверяет частые ошибки с булевыми операторами
//   - buildtag: проверяет корректность build tags
//   - cgocall: проверяет нарушения правил передачи указателей в cgo
//   - composite: проверяет неизящные составные литералы
//   - copylock: проверяет копирование блокировок
//   - errorsas: проверяет правильность использования errors.As
//   - httpresponse: проверяет правильность использования HTTP response
//   - loopclosure: проверяет проблемы с замыканиями в циклах
//   - lostcancel: проверяет отмену context.CancelFunc
//   - nilfunc: проверяет бесполезные сравнения функций с nil
//   - printf: проверяет соответствие строк формата и аргументов в printf-функциях
//   - shift: проверяет правильность битовых сдвигов
//   - stdmethods: проверяет соответствие сигнатур стандартных методов
//   - structtag: проверяет корректность тегов структур
//   - tests: проверяет частые ошибки в тестах
//   - unmarshal: проверяет правильность передачи указателей в unmarshal-функции
//   - unreachable: проверяет недостижимый код
//   - unsafeptr: проверяет правильность использования unsafe.Pointer
//   - unusedresult: проверяет неиспользованные результаты вызовов функций
//
// Анализаторы staticcheck.io (класс SA):
//   - SA1000-SA1030: различные проверки на возможные баги
//   - SA2000-SA2003: проверки на неправильное использование стандартной библиотеки
//   - SA3000-SA3001: проверки на неправильное использование testing пакета
//   - SA4000-SA4031: различные проверки кода
//   - SA5000-SA5012: проверки на правильность использования стандартной библиотеки
//   - SA6000-SA6005: проверки производительности
//   - SA9000-SA9008: различные проверки
//
// Дополнительные анализаторы staticcheck.io:
//   - ST1000: проверяет корректность комментариев к пакетам
//   - S1000: проверяет возможность упрощения кода
//
// Публичные анализаторы:
//   - bodyclose: проверяет закрытие HTTP response body
//   - errcheck: проверяет обработку ошибок
//
// Собственный анализатор:
//   - osexit: запрещает прямой вызов os.Exit в функции main пакета main
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
)

func main() {
	checks := []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,

		bodyclose.Analyzer,
		errcheck.Analyzer,

		osExitAnalyzer,
	}

	for _, analyzer := range staticcheck.Analyzers {
		checks = append(checks, analyzer.Analyzer)
	}

	for _, analyzer := range stylecheck.Analyzers {
		if analyzer.Analyzer.Name == "ST1000" {
			checks = append(checks, analyzer.Analyzer)
			break
		}
	}

	for _, analyzer := range simple.Analyzers {
		if analyzer.Analyzer.Name == "S1000" {
			checks = append(checks, analyzer.Analyzer)
			break
		}
	}

	multichecker.Main(checks...)
}
