<?php
// Demo for the hibp FrankenPHP extension. Both functions below are native
// PHP functions implemented in Go and compiled into the FrankenPHP binary.

$password = $_GET['password'] ?? 'P@ssw0rd';
$domain = $_GET['domain'] ?? 'adobe.com';

$count = hibp_pwned_password_count($password);
$breaches = json_decode(hibp_breaches($domain), true);

header('Content-Type: text/html; charset=utf-8');
?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Have I Been Pwned — FrankenPHP demo</title>
    <style>
        body { font-family: system-ui, sans-serif; max-width: 46rem; margin: 3rem auto; padding: 0 1rem; }
        code { background: #f2f2f2; padding: .15rem .35rem; border-radius: 4px; }
        .pwned { color: #b00020; } .ok { color: #1b7f3b; } .err { color: #a15c00; }
    </style>
</head>
<body>
<h1>Have I Been Pwned — FrankenPHP demo</h1>
<p>These results come from PHP functions implemented in Go, compiled directly into FrankenPHP.</p>

<h2><code>hibp_pwned_password_count('<?= htmlspecialchars($password) ?>')</code></h2>
<?php if ($count < 0): ?>
    <p class="err">Lookup failed — is the container allowed to reach api.pwnedpasswords.com?</p>
<?php elseif ($count > 0): ?>
    <p class="pwned">⚠️ This password appears <strong><?= number_format($count) ?></strong> times in known breaches. Do not use it.</p>
<?php else: ?>
    <p class="ok">✅ This password was not found in the Pwned Passwords corpus.</p>
<?php endif; ?>

<h2><code>hibp_breaches('<?= htmlspecialchars($domain) ?>')</code></h2>
<?php if (isset($breaches['error'])): ?>
    <p class="err"><?= htmlspecialchars($breaches['error']) ?></p>
<?php elseif (empty($breaches)): ?>
    <p class="ok">No breaches recorded for this domain.</p>
<?php else: ?>
    <ul>
        <?php foreach ($breaches as $breach): ?>
            <li>
                <strong><?= htmlspecialchars($breach['Title'] ?? $breach['Name']) ?></strong>
                (<?= htmlspecialchars($breach['BreachDate'] ?? 'unknown date') ?>)
                — <?= number_format($breach['PwnCount'] ?? 0) ?> accounts
            </li>
        <?php endforeach; ?>
    </ul>
<?php endif; ?>

<p>Try it: <code>?password=correct+horse&amp;domain=linkedin.com</code></p>
</body>
</html>
