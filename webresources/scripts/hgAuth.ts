import { AccountInfo, PublicClientApplication } from "@azure/msal-browser";
import * as utils from "../common/utils";
import * as constants from "../common/constants";

declare const msal: typeof import("@azure/msal-browser");

class HeterogenAuth {
  private msalInstance: PublicClientApplication | null = null;
  private readonly clientId = "bb3c49a3-171c-49ab-a96b-8b247f600c44";
  private readonly tenantId = "09af08d8-4dbc-4605-b414-2c6d9c7a0e70";
  private readonly domain = "https://heterogenportalusers.ciamlogin.com";
  private readonly openIdEndpoint = "v2.0/.well-known/openid-configuration";
  private readonly userFlowName = "Default";
  private readonly signInEndpoint = "sign-in";
  private readonly redirectEndpoint = "sign-in-redirect";

  public async init() {
    try {
      this.msalInstance = new msal.PublicClientApplication({
        auth: {
          clientId: this.clientId,
          authority: `${this.domain}/${this.tenantId}/${this.openIdEndpoint}?p=${this.userFlowName}`,
          redirectUri: `${utils.getAppBase()}/${this.redirectEndpoint}`,
        },
        system: {},
      });

      await this.msalInstance.initialize();
      await this.handleEndpointTriggers();

      this.initialiseSignOut();
      this.initialiseManualRedirectLink();
      this.initialiseManualSignInLink();
    } catch (err: unknown) {
      utils.handleScriptError(err);
    }
  }

  private handleEndpointTriggers = async () => {
    const endpointTriggerMap: Partial<Record<string, () => Promise<void>>> = {
      [this.signInEndpoint]: this.handleSignIn,
      [this.redirectEndpoint]: this.handleRedirectFromEntraId,
    };

    const triggerAction = endpointTriggerMap[this.getEndpoint()];
    if (triggerAction) {
      await triggerAction();
    }
  };

  private handleSignIn = async () => {
    if (this.msalInstance === null) {
      throw new Error("msal instance is not instantiated");
    }

    // flush any pending interactions
    try {
      await this.msalInstance.handleRedirectPromise({
        navigateToLoginRequestUrl: false,
      });
    } catch {}

    const activeAccount =
      this.msalInstance.getActiveAccount() ??
      this.msalInstance.getAllAccounts()[0];

    if (activeAccount) {
      await this.signActiveUserIntoApp(activeAccount);
    } else {
      this.msalInstance
        .loginRedirect()
        .catch((err) => utils.handleScriptError(err));
    }
  };

  private initialiseManualRedirectLink = () => {
    const manualRedirectLink = document.getElementById(
      constants.MANUAL_REDIRECT_LINK_ID,
    );
    if (!manualRedirectLink) {
      return;
    }

    manualRedirectLink.addEventListener("click", this.handleSignIn);
  };

  private initialiseManualSignInLink = () => {
    const manualSignInLink = document.getElementById(
      constants.MANUAL_SIGN_IN_LINK_ID,
    );
    if (!manualSignInLink) {
      return;
    }

    manualSignInLink.addEventListener("click", this.handleRedirectFromEntraId);
  };

  private handleRedirectFromEntraId = async () => {
    if (this.msalInstance === null) {
      throw new Error("msal instance is not instantiated");
    }
    const response = await this.msalInstance.handleRedirectPromise({
      navigateToLoginRequestUrl: false,
    });

    if (!response?.account) {
      utils.handleScriptError(
        new Error("Unable to read response from authentication service"),
      );
      return;
    }

    this.msalInstance.setActiveAccount(response.account);
    await this.signActiveUserIntoApp(response.account);
  };

  private initialiseSignOut = () => {
    document.addEventListener("click", (e) => {
      const target = utils.GetEventTarget(e);
      if (!target) {
        return;
      }

      if (!target.closest("#sign-out-button")) {
        return;
      }

      try {
        if (this.msalInstance === null) {
          throw new Error("msal instance is not instantiated");
        }
        this.msalInstance.logoutRedirect({
          postLogoutRedirectUri: "/signed-out",
        });
      } catch (err: unknown) {
        utils.handleScriptError(err);
      }
    });
  };

  private signActiveUserIntoApp = async (activeUser: AccountInfo) => {
    if (this.msalInstance === null) {
      utils.handleScriptError(new Error("msal instance is not instantiated"));
      return;
    }

    const authResult = await this.msalInstance.acquireTokenSilent({
      account: activeUser,
      scopes: [`api://${this.clientId}/default`],
    });

    const signInButton = document.getElementById("sign-in-button");
    if (!signInButton) {
      throw new Error("Unable to access sign-in button");
    }

    signInButton.setAttribute(
      "hx-headers",
      JSON.stringify({ Authorization: `Bearer ${authResult.accessToken}` }),
    );

    signInButton.click();
  };

  private getEndpoint = () => {
    return window.location.pathname.split("/").filter(Boolean).pop() || "";
  };
}

(async () => {
  try {
    const inst = new HeterogenAuth();
    await inst.init();
    utils.attachToHeterogenNamespace({ HeterogenAuth: inst });
  } catch (err) {
    utils.handleScriptError(err);
  }
})();
