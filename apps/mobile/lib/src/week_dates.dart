DateTime currentWeekStart({DateTime? now}) {
  final today = now ?? DateTime.now();
  final localDate = DateTime(today.year, today.month, today.day);
  return localDate.subtract(
    Duration(days: localDate.weekday - DateTime.monday),
  );
}

String currentWeekTargetDate({DateTime? now}) {
  return formatDateOnly(now ?? DateTime.now());
}

String formatDateOnly(DateTime date) {
  final localDate = DateTime(date.year, date.month, date.day);
  return [
    localDate.year.toString().padLeft(4, '0'),
    localDate.month.toString().padLeft(2, '0'),
    localDate.day.toString().padLeft(2, '0'),
  ].join('-');
}
