/* Hand-written arginfo for the hibp extension (kept in sync with
 * extension.stub.php). */

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_hibp_pwned_password_count, 0, 1, IS_LONG, 0)
	ZEND_ARG_TYPE_INFO(0, password, IS_STRING, 0)
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_WITH_RETURN_TYPE_INFO_EX(arginfo_hibp_breaches, 0, 0, IS_STRING, 0)
	ZEND_ARG_TYPE_INFO_WITH_DEFAULT_VALUE(0, domain, IS_STRING, 0, "\"\"")
ZEND_END_ARG_INFO()

ZEND_FUNCTION(hibp_pwned_password_count);
ZEND_FUNCTION(hibp_breaches);

static const zend_function_entry ext_functions[] = {
	ZEND_FE(hibp_pwned_password_count, arginfo_hibp_pwned_password_count)
	ZEND_FE(hibp_breaches, arginfo_hibp_breaches)
	ZEND_FE_END
};
