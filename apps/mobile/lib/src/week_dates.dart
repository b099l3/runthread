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

DateTime weekStartFor(DateTime date) {
  final localDate = DateTime(date.year, date.month, date.day);
  return localDate.subtract(
    Duration(days: localDate.weekday - DateTime.monday),
  );
}

String compactWeekRange(DateTime startsOn) {
  final end = startsOn.add(const Duration(days: 6));
  if (startsOn.month == end.month) {
    return '${_monthLabel(startsOn.month)} ${startsOn.day}-${end.day}';
  }
  return '${_monthLabel(startsOn.month)} ${startsOn.day}-${_monthLabel(end.month)} ${end.day}';
}

String formatDateOnly(DateTime date) {
  final localDate = DateTime(date.year, date.month, date.day);
  return [
    localDate.year.toString().padLeft(4, '0'),
    localDate.month.toString().padLeft(2, '0'),
    localDate.day.toString().padLeft(2, '0'),
  ].join('-');
}

String _monthLabel(int month) {
  const months = [
    'Jan',
    'Feb',
    'Mar',
    'Apr',
    'May',
    'Jun',
    'Jul',
    'Aug',
    'Sep',
    'Oct',
    'Nov',
    'Dec',
  ];
  if (month < 1 || month > months.length) {
    return '';
  }
  return months[month - 1];
}
