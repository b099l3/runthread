enum DistanceUnit {
  kilometers,
  miles;

  String get label {
    return switch (this) {
      DistanceUnit.kilometers => 'km',
      DistanceUnit.miles => 'mi',
    };
  }

  String get storageValue {
    return switch (this) {
      DistanceUnit.kilometers => 'km',
      DistanceUnit.miles => 'mi',
    };
  }

  double distanceFromMeters(double meters) {
    return switch (this) {
      DistanceUnit.kilometers => meters / 1000,
      DistanceUnit.miles => meters / 1609.344,
    };
  }
}

const distanceUnitPreferenceKey = 'distance_unit';

DistanceUnit distanceUnitFromStorageValue(String? value) {
  return switch (value) {
    'mi' => DistanceUnit.miles,
    _ => DistanceUnit.kilometers,
  };
}

String distanceLabelFromMeters(double meters, DistanceUnit unit) {
  if (meters <= 0) {
    return '';
  }
  final distance = unit.distanceFromMeters(meters);
  return '${distance.toStringAsFixed(distance >= 10 ? 0 : 1)} ${unit.label}';
}
