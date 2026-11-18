// View for the statechart skill: reader-facing labels for UserAttributeUpdate with race.cfg.
// Generate the chart in a scratch folder outside the repo, not in this folder:
//   statechart UserAttributeUpdate.tla --config race.cfg --view statechart-view.mjs --out <new-folder>
// Values us and eu stand for two user lists that a user can set in the Lightdash UI.

const phase = {
  idle: "before the apply",
  read: "apply running",
  done: "after the apply",
};

export const state = ({ pc, serverUsers, lastUI, snapshot }) => {
  if (pc === "idle") return `Before the apply: Lightdash holds ${serverUsers}`;
  if (pc === "read") {
    return serverUsers === snapshot
      ? `Apply read ${snapshot}, Lightdash still holds ${snapshot}`
      : `Apply read ${snapshot}, a UI edit changed Lightdash to ${serverUsers}`;
  }
  return serverUsers === lastUI
    ? `Done: Lightdash holds ${lastUI}`
    : `Done: the UI value ${lastUI} is lost, Lightdash holds ${serverUsers}`;
};

export const step = (action, from) => {
  switch (action) {
    case "UIEdit":
      return from.pc === "read"
        ? "a user sets a new value in Lightdash [while the apply runs]"
        : "a user sets a new value in Lightdash [before the apply starts]";
    case "Get":
      return "the apply reads the current value";
    case "Put":
      return "the apply writes back the value it read";
    case "Done":
      return null;
    default:
      throw new Error(`no label for action ${action}`);
  }
};

export const group = ({ pc }) => phase[pc];

export const colors = {
  "before the apply": "blue",
  "apply running": "orange",
  "after the apply": "teal",
};

export const title =
  "lightdash_user_attribute Update: a UI edit during the apply is lost";
