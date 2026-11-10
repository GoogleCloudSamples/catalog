`PollSettings` requires both an `Expiration` (the 90-second total wait timeout)
and a `delay` between `GetOperation` status checks (`TimeSpan.FromSeconds(10)`
matches `Google.LongRunning`'s default polling interval).
