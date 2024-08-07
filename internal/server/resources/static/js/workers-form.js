(function () {
  function debounce(fn, timeout) {
    let timerId;
    return function (...params) {
      if (timerId) {
        clearTimeout(timerId);
      }
      timerId = setTimeout(fn, timeout, ...params);
    };
  }
  function itemIcon(item) {
    const wrapper = document.createElement("div");
    wrapper.setAttribute("data-tooltip", item.uid + " - " + item.displayName);
    wrapper.classList.add("item-stack");

    const link = document.createElement("a");
    link.setAttribute("href", "/items/" + item.uid);

    const image = document.createElement("img");
    image.setAttribute("src", "data:image/png;base64, " + item.icon);
    image.style.width = "48px";
    image.style.height = "48px";

    link.appendChild(image);

    wrapper.appendChild(link);
    return wrapper;
  }

  // <div data-tooltip={ fmt.Sprintf("%s - %s", item.UID, item.DisplayName) } class="item-stack">
  // 	<a href={ templ.URL(fmt.Sprintf("/items/%s", item.UID)) }>
  // 		<img src={ "data:image/png;base64, " + item.Base64Icon() } style="width: 48px; height: 48px;"/>
  // 	</a>
  // </div>

  let counter = 1;
  customElements.define(
    "exporter-config",
    class extends HTMLElement {
      constructor() {
        super();
      }

      connectedCallback() {
        const idx = counter++;

        const fieldSet = document.createElement("fieldset");
        fieldSet.classList.add("grid");

        const createInput = (title, name, def) => {
          const label = document.createElement("label");
          label.appendChild(document.createTextNode(title));
          const i = document.createElement("input");
          i.setAttribute("type", "text");
          i.classList.add("field-" + name);
          i.setAttribute("name", name + "_" + idx);
          const value = this.getAttribute(name) || def || "";
          i.setAttribute("value", value);
          label.appendChild(i);
          return label;
        };

        fieldSet.appendChild(createInput("Storage for exports", "storage"));
        const diag = document.createElement("item-selector");
        diag.style.display = "block";
        diag.setAttribute("name", "item_" + idx);
        const label = document.createElement("label");
        label.appendChild(document.createTextNode("Item"));
        label.appendChild(diag);

        fieldSet.appendChild(label);
        fieldSet.appendChild(createInput("Slot", "slot"));
        fieldSet.appendChild(createInput("Amount", "amount", "64"));

        const removeBtn = document.createElement("button");
        removeBtn.innerHTML = "Del";
        removeBtn.addEventListener("click", () => {
          this.remove();
        });
        fieldSet.appendChild(removeBtn);
        this.appendChild(fieldSet);
      }
    }
  );

  customElements.define(
    "importer-config",
    class extends HTMLElement {
      constructor() {
        super();
      }

      connectedCallback() {
        const idx = counter++;

        const fieldSet = document.createElement("fieldset");
        fieldSet.classList.add("grid");

        const createInput = (title, name, def) => {
          const label = document.createElement("label");
          label.appendChild(document.createTextNode(title));
          const i = document.createElement("input");
          i.setAttribute("type", "text");
          i.classList.add("field-" + name);
          i.setAttribute("name", name + "_" + idx);
          const value = this.getAttribute(name) || def || "";
          i.setAttribute("value", value);
          label.appendChild(i);
          return label;
        };

        fieldSet.appendChild(createInput("Storage for import", "storage"));
        fieldSet.appendChild(createInput("Slot", "slot"));

        const removeBtn = document.createElement("button");
        removeBtn.innerHTML = "Del";
        removeBtn.addEventListener("click", () => {
          this.remove();
        });
        fieldSet.appendChild(removeBtn);
        this.appendChild(fieldSet);
      }
    }
  );

  customElements.define(
    "exporter-configs",
    class extends HTMLElement {
      constructor() {
        super();

        this.querySelector(".add-worker").addEventListener("click", (e) => {
          e.preventDefault();
          const worker = document.createElement("exporter-config");
          this.appendChild(worker);
        });
      }
    }
  );

  customElements.define(
    "importer-configs",
    class extends HTMLElement {
      constructor() {
        super();

        this.querySelector(".add-worker").addEventListener("click", (e) => {
          e.preventDefault();
          const worker = document.createElement("importer-config");
          this.appendChild(worker);
        });
      }
    }
  );

  /*
<details class="dropdown">
  <summary>Dropdown</summary>
  <ul>
    <li><a href="#">Solid</a></li>
    <li><a href="#">Liquid</a></li>
    <li><a href="#">Gas</a></li>
    <li><a href="#">Plasma</a></li>
  </ul>
</details>
     */

  const itemSelectorTemplate = `
	<dialog open id="item-select">
		<article>
			<h2>Items</h2>
			<form>
				<fieldset>
					<label>
						Name
						<input name="filter" />
					</label>
				</fieldset>
			</form>
			<div id="item-popup-items"></div>
			<footer>
				<button class="cancel-btn">Cancel</button>
			</footer>
		</article>
	</dialog>
    `;

  let dialogElement = null;

  customElements.define(
    "item-select-dialog",
    class extends HTMLElement {
      constructor() {
        super();
        this.addEventListener("show-dialog", (e) => {
          const [resolve, reject] = e.detail;
          this.resolve = resolve;
          this.reject = reject;
          this.style.display = "block";
          this.querySelector("input").focus();
        });
      }
      connectedCallback() {
        this.style.display = "none";
        const fragment = document
          .createRange()
          .createContextualFragment(itemSelectorTemplate);

        this.appendChild(fragment);

        this.querySelector(".cancel-btn").addEventListener("click", () => {
          if (this.reject != null) {
            this.reject();
          }
          this.reject = null;
          this.resolve = null;
          this.style.display = "none";
        });
        const content = this.querySelector("#item-popup-items");
        const input = this.querySelector("input");
        const complete = function (item) {
          if (this.resolve != null) {
            this.resolve(item);
          }
          this.resolve = null;
          this.reject = null;
          this.style.display = "none";
        }.bind(this);
        input.addEventListener(
          "keyup",
          debounce(async () => {
            const response = await fetch(
              "/item-suggest/?" +
                new URLSearchParams({
                  filter: input.value,
                }).toString()
            );
            const data = await response.json();

            const table = document.createElement("table");
            for (let item of data) {
              const row = document.createElement("tr");
              const icon = document.createElement("td");
              icon.appendChild(itemIcon(item));

              row.appendChild(icon);
              const name = document.createElement("td");
              name.innerHTML = item.displayName;
              row.appendChild(name);
              const buttons = document.createElement("td");
              const select = document.createElement("button");
              select.innerHTML = "Select";
              select.addEventListener("click", () => complete(item));
              buttons.appendChild(select);
              row.appendChild(buttons);

              table.appendChild(row);
            }

            content.replaceChildren(table);

            console.log(data);
          }, 200)
        );
      }
    }
  );

  function selectItem() {
    if (dialogElement == null) {
      const dialog = document.createElement("item-select-dialog");
      document.body.appendChild(dialog);
      dialogElement = dialog;
    }
    return new Promise((resolve, reject) => {
      const event = new CustomEvent("show-dialog", {
        detail: [resolve, reject],
      });
      dialogElement.dispatchEvent(event);
    });
  }

  customElements.define(
    "item-selector",
    class extends HTMLElement {
      constructor() {
        super();
      }

      connectedCallback() {
        // const details = document.createElement("details");
        // details.classList.add("dropdown");
        // const summary = document.createElement("summary");
        // const i = document.createElement("input");
        // summary.appendChild(i);

        // const ul = document.createElement("ul");
        // details.appendChild(summary);
        // details.appendChild(ul);

        // const li = document.createElement("li");
        // li.innerHTML = "TEST";
        // ul.appendChild(li);
        const wrap = document.createElement("span");
        this.appendChild(wrap);

        const name = this.getAttribute("name");

        const btn = document.createElement("button");
        btn.innerHTML = "Sel";
        btn.addEventListener("click", (e) => {
          e.preventDefault();
          selectItem().then((item) => {
            const input = document.createElement("input");
            input.setAttribute("type", "hidden");
            input.setAttribute("name", name);
            input.setAttribute("value", item.uid);

            wrap.replaceChildren(itemIcon(item), input);
          });
        });

        this.appendChild(btn);
      }
    }
  );
})();
