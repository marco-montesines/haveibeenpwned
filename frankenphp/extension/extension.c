/* C bridge between the Zend engine and the Go implementation in
 * extension.go. The Go symbols are provided by cgo via _cgo_export.h. */

#include <php.h>
#include "extension.h"
#include "extension_arginfo.h"
#include "_cgo_export.h"

PHP_FUNCTION(hibp_pwned_password_count)
{
	zend_string *password;

	ZEND_PARSE_PARAMETERS_START(1, 1)
		Z_PARAM_STR(password)
	ZEND_PARSE_PARAMETERS_END();

	RETURN_LONG(hibp_pwned_password_count(password));
}

PHP_FUNCTION(hibp_breaches)
{
	zend_string *domain = NULL;

	ZEND_PARSE_PARAMETERS_START(0, 1)
		Z_PARAM_OPTIONAL
		Z_PARAM_STR(domain)
	ZEND_PARSE_PARAMETERS_END();

	zend_string *result = hibp_breaches(domain ? domain : ZSTR_EMPTY_ALLOC());

	RETURN_STR(result);
}

zend_module_entry hibp_module_entry = {
	STANDARD_MODULE_HEADER,
	"hibp",
	ext_functions,
	NULL, /* MINIT */
	NULL, /* MSHUTDOWN */
	NULL, /* RINIT */
	NULL, /* RSHUTDOWN */
	NULL, /* MINFO */
	"1.0.0",
	STANDARD_MODULE_PROPERTIES
};
