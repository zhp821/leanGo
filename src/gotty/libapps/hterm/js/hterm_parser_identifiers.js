// Copyright (c) 2015 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

'use strict';

/**
 * Collections of identifier for hterm.Parser.
 */
hterm.Parser.identifiers = {};

hterm.Parser.identifiers.modifierKeys = {
  Shift: 'shift',
  Ctrl: 'ctrl',
  Alt: 'alt',
  Meta: 'meta'
};

/**
 * Key codes useful when defining key sequences.
 *
 * Punctuation is mostly left out of this list because they can move around
 * based on keyboard locale and browser.
 *
 * In a key sequence like "Ctrl-ESC", the ESC comes from this list of
 * identifiers.  It is equivalent to "Ctrl-27" and "Ctrl-0x1b".
 */
hterm.Parser.identifiers.keyCodes = {
  // Top row.
  ESC: 27,
  F1: 112,
  F2: 113,
  F3: 114,
  F4: 115,
  F5: 116,
  F6: 117,
  F7: 118,
  F8: 119,
  F9: 120,
  F10: 121,
  F11: 122,
  F12: 123,

  // Row two.
  ONE: 49,
  TWO: 50,
  THREE: 51,
  FOUR: 52,
  FIVE: 53,
  SIX: 54,
  SEVEN: 55,
  EIGHT: 56,
  NINE: 57,
  ZERO: 48,
  BACKSPACE: 8,

  // Row three.
  TAB: 9,
  Q: 81,
  W: 87,
  E: 69,
  R: 82,
  T: 84,
  Y: 89,
  U: 85,
  I: 73,
  O: 79,
  P: 80,

  // Row four.
  CAPSLOCK: 20,
  A: 65,
  S: 83,
  D: 68,
  F: 70,
  G: 71,
  H: 72,
  J: 74,
  K: 75,
  L: 76,
  ENTER: 13,

  // Row five.
  Z: 90,
  X: 88,
  C: 67,
  V: 86,
  B: 66,
  N: 78,
  M: 77,

  // Etc.
  SPACE: 32,
  PRINT_SCREEN: 42,
  SCROLL_LOCK: 145,
  BREAK: 19,
  INSERT: 45,
  HOME: 36,
  PGUP: 33,
  DEL: 46,
  END: 35,
  PGDOWN: 34,
  UP: 38,
  DOWN: 40,
  RIGHT: 39,
  LEFT: 37,
  NUMLOCK: 144,

  // Keypad
  KP0: 96,
  KP1: 97,
  KP2: 98,
  KP3: 99,
  KP4: 100,
  KP5: 101,
  KP6: 102,
  KP7: 103,
  KP8: 104,
  KP9: 105,
  KP_PLUS: 107,
  KP_MINUS: 109,
  KP_STAR: 106,
  KP_DIVIDE: 111,
  KP_DECIMAL: 110,

  // Chrome OS media keys
  NAVIGATE_BACK: 166,
  NAVIGATE_FORWARD: 167,
  RELOAD: 168,
  FULL_SCREEN: 183,
  WINDOW_OVERVIEW: 182,
  BRIGHTNESS_UP: 216,
  BRIGHTNESS_DOWN: 217
};

/**
 * Identifiers for use in key actions.
 */
hterm.Parser.identifiers.actions = {
  /**
   * Prevent the browser and operating system from handling the event.
   */
  CANCEL: hterm.Keyboard.KeyActions.CANCEL,

  /**
   * Wait for a "keypress" event, send the keypress charCode to the host.
   */
  DEFAULT: hterm.Keyboard.KeyActions.DEFAULT,

  /**
   * Let the browser or operating system handle the key.
   */
  PASS: hterm.Keyboard.KeyActions.PASS,

  /**
   * Scroll the terminal one page up.
   */
  scrollPageUp: function(terminal) {
    terminal.scrollPageUp();
    return hterm.Keyboard.KeyActions.CANCEL;
  },

  /**
   * Scroll the terminal one page down.
   */
  scrollPageDown: function(terminal) {
    terminal.scrollPageDown();
    return hterm.Keyboard.KeyActions.CANCEL;
  },

  /**
   * Scroll the terminal to the top.
   */
  scrollToTop: function(terminal) {
    terminal.scrollEnd();
    return hterm.Keyboard.KeyActions.CANCEL;
  },

  /**
   * Scroll the terminal to the bottom.
   */
  scrollToBottom: function(terminal) {
    terminal.scrollEnd();
    return hterm.Keyboard.KeyActions.CANCEL;
  },

  /**
   * Clear the terminal and scrollback buffer.
   */
  clearScrollback: function(terminal) {
    terminal.wipeContents();
    return hterm.Keyboard.KeyActions.CANCEL;
  }
};
