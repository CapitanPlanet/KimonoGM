window.JanuszAudio = {
    musicAudio: null,
    voiceAudio: null,
    sfxAudio: null,
    musicBaseVol: 0.5,
    voiceBaseVol: 0.5,
    sfxBaseVol: 0.5,
    duckingLevel: 0.15,
    isVoiceMuted: false,
    _currentMusicSrc: null,

    // FIX 1: normalizacja sciezki. Edytor zapisuje teraz "sfx/plik",
    // stare zapisy maja "sounds/sfx/plik", a JS wczesniej bral to jak leci.
    // Zawsze zwracamy "/sounds/sfx/plik" - jedno sounds/, nigdy dwa.
    resolveSrc: function(src, kind) {
        if (!src) return null;
        kind = kind || "sfx";
        var s = String(src).replace(/\\/g, "/").trim();
        if (s.indexOf("data:") === 0 || s.indexOf("http") === 0 || s.indexOf("blob:") === 0) return s;
        // zdejmij wszystkie wiodace sounds/ i /sounds/
        while (s.indexOf("sounds/") === 0) s = s.slice(7);
        while (s.indexOf("/sounds/") === 0) s = s.slice(8);
        while (s.indexOf("./") === 0) s = s.slice(2);
        // gola nazwa -> sfx/nazwa, voice/nazwa, music/nazwa
        if (s.indexOf("/") === -1) s = kind + "/" + s;
        // jak juz ma sfx/, voice/, music/ to nic nie doklejamy, jak nie ma to doklejamy kind
        if (s.indexOf("sfx/") !== 0 && s.indexOf("voice/") !== 0 && s.indexOf("music/") !== 0) s = kind + "/" + s;
        return "/sounds/" + s;
    },

    // FIX 2: JEDNA funkcja na POCZATEK sceny.
    // Silnik ma ja wolac w renderScene(), PRZED pokazaniem tekstu, a nie po kliknieciu dalej.
    //   JanuszAudio.playSceneAudio(scene)
    // gdzie scene = { music: "music/plik", voice: "voice/plik", sfx: "sfx/plik" }
    // albo scene = { audio: "sfx/plik" } / scene = { sound: "sfx/plik" }
    // Puste pola = nic nie gra, bez bledu.
    playSceneAudio: function(scene) {
        if (!scene) return;
        // Zatrzymaj poprzedni voice/sfx, zeby nie lecial stary dzwiek w nowej scenie
        this.stopVoice();
        this.stopSfx();
        // Muzyka: jak ta sama co poprzednio, nie restartuj jej
        var music = scene.music || scene.musicFile || scene.bgMusic || null;
        if (music) this.startMusic(music);
        // Voice i SFX: graja OD RAZU na poczatku sceny, nie po kliknieciu
        var voice = scene.voice || scene.voiceFile || null;
        if (voice) this.playVoice(voice);
        var sfx = scene.sfx || scene.sound || scene.audio || scene.sfxFile || null;
        if (sfx) this.playSfx(sfx);
    },
    // Alias, jakby silnik wolal inna nazwe
    playSceneSounds: function(scene) { this.playSceneAudio(scene); },
    onSceneStart: function(scene) { this.playSceneAudio(scene); },

    stopVoice: function() {
        if (this.voiceAudio) { try { this.voiceAudio.pause(); } catch(e){} this.voiceAudio.currentTime = 0; }
    },
    stopSfx: function() {
        if (this.sfxAudio) { try { this.sfxAudio.pause(); } catch(e){} this.sfxAudio.currentTime = 0; }
        // playSfx uzywa voiceAudio (stary silnik), wiec stopSfx tez go zatrzymuje jak to byl sfx
    },
    stopSceneAudio: function() { this.stopVoice(); this.stopSfx(); },

    setMusicVol: function(vol) {
        this.musicBaseVol = parseFloat(vol);
        if (this.musicAudio) {
            if (this.voiceAudio && !this.voiceAudio.paused && !this.voiceAudio.ended && !this.isVoiceMuted) {
                this.musicAudio.volume = this.duckingLevel;
            } else {
                this.musicAudio.volume = this.musicBaseVol;
            }
        }
    },
    // ALIASY dla starego silnika
    setSfxVol: function(vol) { this.setVoiceVol(vol); },
    setVoiceVolume: function(vol) { this.setVoiceVol(vol); },

    setVoiceVol: function(vol) {
        this.voiceBaseVol = parseFloat(vol);
        this.sfxBaseVol = parseFloat(vol);
        this.isVoiceMuted = vol == 0;
        if (this.voiceAudio) this.voiceAudio.volume = this.voiceBaseVol;
        if (this.sfxAudio) this.sfxAudio.volume = this.sfxBaseVol;
        if (this.isVoiceMuted && this.musicAudio) this.musicAudio.volume = this.musicBaseVol;
        else if (!this.isVoiceMuted && this.voiceAudio && !this.voiceAudio.paused && !this.voiceAudio.ended) {
            if (this.musicAudio) this.musicAudio.volume = this.duckingLevel;
        }
    },

    startMusic: function(src) {
        var url = this.resolveSrc(src, "music");
        if (!url) return;
        // Jak ta sama muzyka juz gra, nie restartuj jej przy zmianie sceny
        if (this.musicAudio && this._currentMusicSrc === url && !this.musicAudio.paused) return;
        this._currentMusicSrc = url;
        if (!this.musicAudio) {
            this.musicAudio = new Audio(url);
            this.musicAudio.loop = true;
            this.musicAudio.volume = this.musicBaseVol;
            this.musicAudio.play().catch(e => console.log("Muzyka czeka na klik:", e));
        } else {
            this.musicAudio.src = url;
            this.musicAudio.loop = true;
            this.musicAudio.play().catch(()=>{});
        }
    },
    playMusic: function(src) { this.startMusic(src); },

    playVoice: function(src) {
        var url = this.resolveSrc(src, "voice");
        if (!url) return;
        if (this.voiceAudio) { this.voiceAudio.pause(); this.voiceAudio.currentTime = 0; }
        this.voiceAudio = new Audio(url);
        this.voiceAudio.volume = this.voiceBaseVol;
        this.voiceAudio.addEventListener('play', () => {
            if (!this.isVoiceMuted && this.musicAudio && !this.musicAudio.paused) this.musicAudio.volume = this.duckingLevel;
        });
        const restoreMusic = () => { if (this.musicAudio && !this.musicAudio.paused) this.musicAudio.volume = this.musicBaseVol; };
        this.voiceAudio.addEventListener('ended', restoreMusic);
        this.voiceAudio.addEventListener('pause', restoreMusic);
        if (!this.isVoiceMuted) this.voiceAudio.play().catch(e => console.log("Voice error:", e));
    },
    playSfx: function(src) { this.playVoice(src); },
    playSound: function(src) { this.playVoice(src); },
    stopMusic: function() { if(this.musicAudio) this.musicAudio.pause(); this._currentMusicSrc = null; }
};
