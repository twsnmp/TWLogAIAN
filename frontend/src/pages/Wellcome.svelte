<script>
  import logo from "../assets/images/appicon.png"  
  import { MortarBoard16, PaperAirplane16, Sun16, Moon16,ThreeBars16 } from "svelte-octicons";
  import { createEventDispatcher,onMount } from "svelte";
  import { _,setLocale,getLocale } from '../i18n/i18n';
  import {GetVersion,GetDark,SetDark} from '../../wailsjs/go/main/App';
  import { BrowserOpenURL } from "../../wailsjs/runtime/runtime.js";

  const dispatch = createEventDispatcher();
  let version = "1.0.0(xxxxx)";

  const feedback = () => {
    dispatch("done", { page: "feedback" });
  };
  const workdir = () => {
    dispatch("done", { page: "workdir" });
  };

  let dark = true;
  
  onMount(async () => {
    version = await GetVersion();
    const v = await GetDark();
    if (dark != v) {
      dark = v;
      toggleDark()
    }
  });

  const toggleDark = () => {
    dark = !dark;
    SetDark(dark)
    const e = document.querySelector("body");
    if (e) {
      e.dataset.colorMode = dark ? "dark" : "light";
      e.dataset.darkTheme = dark ? "dark" : "light";
    }
  };
  let locale = getLocale();
  const _setLocale = (l) => {
    locale = l;
    setLocale(l);
  }
</script>

<div class="Box mx-auto" style="max-width: 800px;">
  <div class="Box-header">
    <h3 class="Box-title">{$_('Wellcome.Title')}
      <button class="btn float-right" type="button" on:click={toggleDark}>
        {#if dark}
          <Sun16 />
        {:else}
          <Moon16 />
       {/if}
      </button>
      <details class="details-reset details-overlay float-right">
        <summary class="btn" aria-haspopup="true">
          <ThreeBars16/>
        </summary>
        <div class="SelectMenu">
          <div class="SelectMenu-modal">
            <div class="SelectMenu-list">
              <button class="SelectMenu-item" role="menuitem" on:click={()=>_setLocale('en')} disabled={locale =='en'}>en</button>
              <button class="SelectMenu-item" role="menuitem" on:click={()=>_setLocale('ja')} disabled={locale =='ja'}>ja</button>
            </div>
          </div>
        </div>
      </details>
      </h3>
  </div>
  <div class="Box-body">
    <div class="mx-auto" style="max-width: 200px;">
      <img id="logo" alt="TWLogAIAN Logo" src={logo} />
    </div>
    <hr />
    <p>
      {$_('Wellcome.Line1')}<br />
      {$_('Wellcome.Line2')}<br />
      <a
        href="##"
        on:click={() => {
          BrowserOpenURL("https://note.com/twsnmp/m/m9c88e79743b6");
        }}
      >Note</a><br />
      {$_('Wellcome.Line3')}<br />
      <a
        href="##"
        on:click={() => {
          BrowserOpenURL("https://github.com/twsnmp/TWLogAIAN");
        }}
      >
        GitHUB
      </a><br />
      {$_('Wellcome.Line4')}<br />
      <a
        href="##"
        on:click={() => {
          BrowserOpenURL("https://github.com/twsnmp/TWLogAIAN/issues");
        }}
      >
        GitHub Issues
      </a>
    </p>
    <hr />
    <p>{$_('Wellcome.Thanks')}</p>
    <div class="f6">
      <em><small>TWLogAIAN {version}© 2022 Masayuki Yamai</small></em>
    </div>
  </div>
  <div class="Box-footer text-right">
    <button class="btn btn-danger mr-1" type="button" on:click={feedback}>
      <PaperAirplane16 />
      {$_('Wellcome.Feedback')}
    </button>
    <button class="btn btn-primary" type="button" on:click={workdir}>
      <MortarBoard16 />
      {$_('Wellcome.Start')}
    </button>
  </div>
</div>

<style>
  #logo {
    height: 200px;
    width: 200px;
    transform: rotateY(560deg);
    animation: turn 3.5s ease-out forwards 1s;
  }

  @keyframes turn {
    100% {
      transform: rotateY(0deg);
    }
  }
</style>
