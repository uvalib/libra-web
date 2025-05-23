import { definePreset } from '@primeuix/themes'
import Aura from '@primeuix/themes/aura'
import ripple from '@primeuix/themes/aura/ripple'
import tooltip from '@primeuix/themes/aura/tooltip'
import './styles.scss'
import './styleoverrides.scss'
import colors from './colors.module.scss'

const UVA = definePreset(Aura, {
   root: {
      borderRadius: {
         none: '0',
         xs: '2px',
         sm: '3px',
         md: '4px',
         lg: '4px',
         xl: '8px'
      },
   },
   semantic: {
      primary: {
         50:  colors.brandBlue300,
         100: colors.brandBlue300,
         200: colors.brandBlue300,
         300: colors.brandBlue300,
         400: colors.brandBlue100,
         500: colors.brandBlue100,
         600: colors.brandBlue100,
         700: colors.brandBlue100,
         800: colors.brandBlue,
         900: colors.brandBlue,
         950: colors.brandBlue
      },
      focusRing: {
         width: '2px',
         style: 'dotted',
         offset: '3px'
      },
      formField: {
         paddingX: '0.75rem',
         paddingY: '0.5rem',
         borderRadius: '0.3rem',
         focusRing: {
             width: '1px',
             style: 'dashed',
             color: '{primary.color}',
             offset: '3px',
             shadow: 'none'
         },
      },
      disabledOpacity: '0.3',
      colorScheme: {
         light: {
            primary: {
               color: '{primary.500}',
               contrastColor: '#ffffff',
               hoverColor: '{primary.100}',
               activeColor: '{primary.500}'
            },
            highlight: {
               background: '#ffffff',
               focusBackground: '#ffffff',
               color: colors.textDark,
               focusColor: '#ffffff'
            }
         },
      }
   },
   components: {
      accordion: {
         header: {
            background: '#f8f9fa',
            hoverBackground: '#f5f5ff',
            activeBackground: '#f8f9fa',
            activeHoverBackground: '#f8f9fa',
            borderRadius: 0,
         },
         panel: {
            borderWidth: '1px',
            borderColor: colors.grey200,
            hoverBackground: colors.grey200,
        },
        content: {
            background: '#ffffff',
            borderWidth: '1px 0 0 0',
            padding: '1.125rem 1.125rem 1.125rem 1.125rem'
        }
      },
      button: {
         root: {
            paddingY: '.5em',
            paddingX: '1em',
            gap: '1rem',
            borderRadius: '0.3rem',
            sm: {
               fontSize: '0.875rem',
               paddingX: '0.625rem',
               paddingY: '0.375rem'
           },
           lg: {
               fontSize: '1.5rem',
               paddingX: '1.2rem',
               paddingY: '0.6rem'
           },
         },
         colorScheme: {
            light: {
               success: {
                  background: colors.greenDark,
                  hoverBackground: colors.green,
               },
               secondary: {
                  background: colors.grey200,
                  hoverBackground: colors.grey100,
                  hoverBorderColor: colors.grey,
                  borderColor: colors.grey100,
                  color: colors.textDark,
                  focusRing: {
                     color: colors.brandBlue100,
                  },
               },
               danger: {
                  background: colors.redA,
                  hoverBackground: colors.red,
                  hoverBorderColor: colors.red,
                  borderColor: colors.red,
                  color: "white",
               },
               contrast: {
                  background: colors.brandOrangeDark,
                  hoverBackground: colors.brandOrange,
                  activeBackground: colors.brandOrange,
                  focusRing: {
                     color: 'white',
                     shadow: 'none'
                  }
               },
               info: {
                  background: colors.blueAlt300,
                  activeBackground: colors.blueAlt300,
                  activeColor: '#000000',
                  hoverBackground: '#91d8f2',
                  hoverBorderColor: '#007BAC',
                  borderColor: '#007BAC',
                  color: '#000000',
                  hoverColor: '#000000',
                  borderWidth: '2px'
               },
               text: {
                  primary: {
                     hoverBackground: colors.grey200,
                     activeBackground: colors.grey200,
                     color: colors.textDark,
                  },
               },
               link: {
                  color: colors.linkBase,
               },
               outlined: {
                  secondary: {
                      hoverBackground: colors.grey200,
                      borderColor: colors.grey100,
                      color: colors.textDark,

                  },
               }
            }
         }
      },
      card: {
         root: {
            background: '{content.background}',
            borderRadius: '0.5rem',
            color: colors.textBase,
            shadow: '0 2px 1px -1px #0003, 0 1px 1px #00000024, 0 1px 3px #0000001f',
         },
         title: {
            fontSize: '1.15rem',
            fontWeight: '600'
         }
      },
      datatable: {
         headerCell: {
            borderColor: colors.grey100,
            color: colors.textBase,
            background: colors.grey200,
         },
         bodyCell: {
            borderColor: colors.grey100,
            color: colors.textBase,
         },
      },
      dialog: {
         root: {
            background: '#ffffff',
            borderColor: colors.grey,
            borderRadius: '0.3rem',
         },
         header: {
            padding: '5px 10px',
         },
         content: {
            padding: '1.5rem'
         },
         title: {
            fontWeight: '600',
            fontSize: '1em',
         },
         footer: {
            gap: '1rem'
         }
      },
      menubar: {
         root: {
            borderRadius: "0px",
            background: colors.grey200,
            borderColor: colors.grey200,
            padding: '0.5rem',
         },
         baseItem: {
            borderRadius: '0.3rem',
            padding: '0.5rem 0.75rem',
         },
         item: {
            color: colors.textDark,
            focusBackground: colors.brandBlue200,
            activeBackground: colors.grey100,
            focusColor: 'white',
            activeColor: colors.textBase,
         },
         submenu: {
            background: 'white',
            borderColor: colors.grey100,
            borderRadius: '0.3rem',
            shadow: '{overlay.navigation.shadow}',
            mobileIndent: '1rem',
        },
      },
      panel: {
         root: {
            background: 'white',
            borderColor: colors.grey100,
            color: colors.textBase,
            borderRadius: '0.3rem'
         },
         header: {
            background: colors.grey200,
            borderRadius: '0.3rem 0.3rem 0 0',
            padding: '1rem',
            borderColor: colors.grey100,
            borderWidth: '0px 0px 1px 0px',
         },
         title: {
            fontWeight: '600',
         },
         content: {
            padding: '1.25rem 1.25rem 1.25rem 1.25rem'
         }
      },
      radiobutton: {
         icon: {
            checkedColor: colors.brandBlue300,
         }
      },
      select: {
         root: {
            paddingY: '.5em',
            paddingX: '.5em',
            disabledBackground: '#fafafa',
            disabledColor: '#cacaca',
         },
         option: {
            selectedFocusBackground: colors.blueAlt300,
            selectedFocusColor: colors.textDark,
            selectedBackground: colors.blueAlt300,
            selectedColor: colors.textDark
         }
      },
      toast: {
         root: {
            borderWidth: '1px'
         },
         content: {
            gap: '1rem'
         },
         text: {
            gap: '0.5rem'
         },
         summary: {
            fontWeight: '400',
            fontSize: '1.15rem'
         },
         icon: {
            size: '2rem'
         },
         colorScheme: {
            light: {
               success: {
                  background: colors.green200,
                  borderColor: colors.green,
                  color: colors.textDark
               }
            }
         }
      }
   },
   directives: {
      tooltip,
      ripple
   }
});

export default UVA;