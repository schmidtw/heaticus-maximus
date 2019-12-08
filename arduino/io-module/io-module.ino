//#include <Serial.h>

// The constant that defines how long to debounce the signals.  In micro-seconds.
const long DEBOUNCE_TIME_MS = 20;

const long OUTPUT_REPORT_MIN_INTERVAL_S = 10;

byte _last_state[6] = {0, 0, 0, 0, 0, 0};
unsigned long _pulse_count[6] = { 0UL, 0UL, 0UL, 0UL, 0UL, 0UL };
unsigned long _last_delta_s[6] = { 0UL, 0UL, 0UL, 0UL, 0UL, 0UL };
unsigned long _last_delta_ms[6] = { 0UL, 0UL, 0UL, 0UL, 0UL, 0UL };
unsigned long _seconds_since_boot = 0;
long _output_state = 0;

void setup() {

  /* Setup the outputs and set them to 0 / off. */
  pinMode(8, OUTPUT);
  pinMode(9, OUTPUT);
  pinMode(10, OUTPUT);
  pinMode(11, OUTPUT);
  pinMode(12, OUTPUT);
  pinMode(13, OUTPUT);

  SetRelayState(0);

  /* Setup the inputs. */
  pinMode(7, INPUT);
  pinMode(6, INPUT);
  pinMode(5, INPUT);
  pinMode(4, INPUT);
  pinMode(3, INPUT);
  pinMode(2, INPUT);


  /* Start the serial port */
  Serial.begin(115200) ;
  //while (!Serial) ;
}

void loop() {
  if (true == UpdateInputState()) {
    OutputData();
  }

  while (Serial.available() > 0) {
    int cmd = Serial.read();

    switch (cmd) {
      case '?':
        Serial.println(F("s [0-63] sets the relay output bitmask\ng gets the latest output\n"));
        break;
      case 's':
        _output_state = Serial.parseInt();
        SetRelayState(_output_state);
        break;
      case 'g':
        OutputData();
        break;

    }
  }
}

void OutputData()
{
  Serial.print(F("{"));
  Serial.print( F(" \"SerialNumber\": \"0000001\""));
  Serial.print(F(", \"FirmwareVersion\": \"1.0.0\""));
  Serial.print(F(", \"UpTime\": ")); Serial.print(_seconds_since_boot, DEC);
  Serial.print(F(", \"RelayState\": ")); Serial.print(_output_state, DEC);
  Serial.print(F(", \"Inputs\": {"));
  int j = 0;
  char comma = ' ';
  for (int i = 0; i < 6; i++) {
    Serial.print(comma);
    Serial.print(F("\"")); Serial.print((2 + i), DEC); Serial.print(F("\": {"));
    Serial.print(F("  \"PulseCount\": ")); Serial.print(_pulse_count[i], DEC);
    Serial.print(F(", \"State\": " )); Serial.print(_last_state[i], DEC);
    Serial.print(F(", \"LastPulseTime\": ")); Serial.print(_last_delta_s[i]); Serial.print(F("."));
    if ( 100 > _last_delta_ms[i] ) {
      Serial.print(F("0"));
    }
    if ( 10 > _last_delta_ms[i] ) {
      Serial.print(F("0"));
    }
    Serial.print(_last_delta_ms[i], DEC);
    Serial.print(F(" }"));
    comma = ',';
  }
  Serial.print(F(" } }\n"));
}

void SetRelayState( long out )
{
  int j = 8;
  if ( 0 <= out && out < 64 ) {
    for (int i = 0; i < 6; i++, j++) {
      if ((1 << i) & out) {
        digitalWrite(j, HIGH);
      } else {
        digitalWrite(j, LOW);
      }
    }
  } else {
    Serial.print( "Invalid range.  Expecting: [0-63]\n" );
  }
}

// Figure out the state of the inputs and the clock.
// Return if the output report should be sent.
bool UpdateInputState()
{
  static long debounce_time[6] = { 0L, 0L, 0L, 0L, 0L, 0L };
  static int last_state[6] = { LOW, LOW, LOW, LOW, LOW, LOW };
  static unsigned long last_second = 0;
  static long last_report_s = 0;
  bool rv = false;

  unsigned long now = millis();
  unsigned long ms = now - last_second;

  if (ms > 1000UL) {
    last_second = now;
    _seconds_since_boot++;
    ms -= 1000UL;
  }

  for (int i = 0; i < 6; i++) {
    int io = digitalRead(i + 2);
    if ( io != last_state[i] && now > debounce_time[i]) {
      debounce_time[i] = now + DEBOUNCE_TIME_MS;
      last_state[i] = io;
      _last_state[i] = 1;
      if (LOW == io) {
        _last_state[i] = 0;
      }
      _pulse_count[i]++;
      _last_delta_s[i] = _seconds_since_boot;
      _last_delta_ms[i] = ms;
      rv = true;
    }
  }

  // Always send a report at the minimum interval.
  if (last_report_s + OUTPUT_REPORT_MIN_INTERVAL_S <= _seconds_since_boot ) {
    last_report_s = _seconds_since_boot;
    rv  = true;
  }

  return rv;
}

